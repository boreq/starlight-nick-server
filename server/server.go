package server

import (
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/NYTimes/gziphandler"
	"github.com/boreq/starlight-nick-server/data"
	"github.com/boreq/starlight-nick-server/logging"
	"github.com/boreq/starlight-nick-server/server/api"
	"github.com/boreq/starlight/network/node"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
)

var log = logging.New("server")

type Repository interface {
	// List returns a list of all previously stored nick datas.
	List() ([]data.NickData, error)

	// Put stores nick data which can later be retrieved using the Get
	// method.
	Put(*data.NickData) error

	// Get returns previously stored nick data. If the data is missing nil
	// is returned.
	Get(node.ID) (*data.NickData, error)
}

func Serve(repository Repository, address string) error {
	handler, err := newHandler(repository)
	if err != nil {
		return err
	}

	// Add CORS middleware
	handler = cors.AllowAll().Handler(handler)

	// Add GZIP middleware
	handler = gziphandler.GzipHandler(handler)

	log.Info("starting listening", "address", address)
	return http.ListenAndServe(address, handler)
}

func newHandler(repository Repository) (http.Handler, error) {
	h := &handler{
		repository: repository,
	}

	router := httprouter.New()
	router.GET("/nicks", api.Wrap(h.GetNicks))
	router.PUT("/nicks", api.Wrap(h.PutNick))
	router.GET("/nicks/:id", api.Wrap(h.GetNick))
	return router, nil
}

type handler struct {
	repository Repository
}

func (h *handler) GetNicks(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	nicks, err := h.repository.List()
	if err != nil {
		log.Error("list failed", "err", err)
		return nil, api.InternalServerError
	}
	return nicks, nil
}

func (h *handler) GetNick(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	nodeId, err := hex.DecodeString(getParamString(ps, "id"))
	if err != nil {
		return nil, api.BadRequest.WithMessage("Invalid node ID.")
	}
	nickData, err := h.repository.Get(nodeId)
	if err != nil {
		if isClientError(err) {
			return nil, api.BadRequest.WithMessage(err.Error())
		} else {
			log.Error("get nick failed", "err", err)
			return nil, api.InternalServerError
		}
	}
	if nickData == nil {
		return nil, api.NotFound
	}
	return nickData, nil
}

func (h *handler) PutNick(r *http.Request, ps httprouter.Params) (interface{}, api.Error) {
	if r.Body == nil {
		return nil, api.BadRequest
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, api.BadRequest
	}

	nickData := &data.NickData{}
	if err := json.Unmarshal(body, nickData); err != nil {
		return nil, api.BadRequest
	}

	if err := h.repository.Put(nickData); err != nil {
		if isClientError(err) {
			return nil, api.BadRequest.WithMessage(err.Error())
		} else {
			log.Error("put nick failed", "err", err)
			return nil, api.InternalServerError
		}
	}

	return nil, nil
}

func isClientError(err error) bool {
	return err == data.InvalidNickDataErr ||
		err == data.NewerNickDataPresentErr ||
		err == data.NickConflictErr ||
		err == data.InvalidNodeIdErr
}

func getParamString(ps httprouter.Params, name string) string {
	return strings.TrimSuffix(ps.ByName(name), ".json")
}
