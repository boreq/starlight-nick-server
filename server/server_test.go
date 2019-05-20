package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/boreq/starlight-nick-server/data"
	"github.com/boreq/starlight/network/node"
	"github.com/stretchr/testify/require"
)

type repositoryMock struct {
	listReturn []data.NickData
	listErr    error

	putArgument *data.NickData
	putErr      error

	getArgument *node.ID
	getReturn   *data.NickData
	getErr      error
}

func (r *repositoryMock) List() ([]data.NickData, error) {
	return r.listReturn, r.listErr
}

func (r *repositoryMock) Put(nickData *data.NickData) error {
	r.putArgument = nickData
	return r.putErr
}

func (r *repositoryMock) Get(nodeId node.ID) (*data.NickData, error) {
	r.getArgument = &nodeId
	return r.getReturn, r.getErr
}

func makeComponents(t *testing.T) (*repositoryMock, http.Handler, *httptest.ResponseRecorder) {
	repo := &repositoryMock{}

	h, err := newHandler(repo)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()

	return repo, h, rr
}

func makeNickData() *data.NickData {
	return &data.NickData{
		Id:        []byte("id"),
		Nick:      "nick",
		Time:      time.Date(1990, 1, 1, 1, 1, 1, 1, time.UTC),
		PublicKey: []byte("public key"),
		Signature: []byte("signature"),
	}
}

func makeJsonNickData(t *testing.T) []byte {
	nickData := makeNickData()
	j, err := json.Marshal(nickData)
	if err != nil {
		t.Fatal(err)
	}
	return j
}

func TestGetInvalidNodeId(t *testing.T) {
	// given
	_, h, rr := makeComponents(t)

	req, err := http.NewRequest("GET", "/nicks/jfka", nil)
	if err != nil {
		t.Fatal(err)
	}

	// when
	h.ServeHTTP(rr, req)

	// then
	require.Equal(t, 400, rr.Code, "http status should be Bad Request")
}

func TestGetNonexistent(t *testing.T) {
	// given
	_, h, rr := makeComponents(t)

	req, err := http.NewRequest("GET", "/nicks/abcd", nil)
	if err != nil {
		t.Fatal(err)
	}

	// when
	h.ServeHTTP(rr, req)

	// then
	require.Equal(t, 404, rr.Code, "http status should be Not Found")
}

func TestGet(t *testing.T) {
	// given
	repo, h, rr := makeComponents(t)

	repo.getReturn = makeNickData()

	req, err := http.NewRequest("GET", "/nicks/abcd", nil)
	if err != nil {
		t.Fatal(err)
	}

	// when
	h.ServeHTTP(rr, req)

	// then
	expectedBody := `{"id":"6964","nick":"nick","time":"1990-01-01T01:01:01.000000001Z","publicKey":"cHVibGljIGtleQ==","signature":"c2lnbmF0dXJl"}`
	require.Equal(t, 200, rr.Code, "http status should be OK")
	require.Equal(t, expectedBody, rr.Body.String(), "body should contain json formatted nick data")
}

func TestGetInvalidNodeIdError(t *testing.T) {
	// given
	repo, h, rr := makeComponents(t)

	repo.getErr = data.InvalidNickDataErr

	req, err := http.NewRequest("GET", "/nicks/abcd", nil)
	if err != nil {
		t.Fatal(err)
	}

	// when
	h.ServeHTTP(rr, req)

	// then
	require.Equal(t, 400, rr.Code, "http status should be Bad Request")
}

func TestPut(t *testing.T) {
	// given
	_, h, rr := makeComponents(t)

	buf := bytes.NewBuffer(makeJsonNickData(t))

	req, err := http.NewRequest("PUT", "/nicks", buf)
	if err != nil {
		t.Fatal(err)
	}

	// when
	h.ServeHTTP(rr, req)

	// then
	require.Equal(t, 200, rr.Code, "http status should be OK")
}

func TestPutMalformedJson(t *testing.T) {
	// given
	_, h, rr := makeComponents(t)

	buf := &bytes.Buffer{}
	buf.WriteString("[]")

	req, err := http.NewRequest("PUT", "/nicks", buf)
	if err != nil {
		t.Fatal(err)
	}

	// when
	h.ServeHTTP(rr, req)

	// then
	require.Equal(t, 400, rr.Code, "http status should be Bad Request")
}

func TestPutNoBody(t *testing.T) {
	// given
	_, h, rr := makeComponents(t)

	req, err := http.NewRequest("PUT", "/nicks", nil)
	if err != nil {
		t.Fatal(err)
	}

	// when
	h.ServeHTTP(rr, req)

	// then
	require.Equal(t, 400, rr.Code, "http status should be Bad Request")
}

func TestPutClientErr(t *testing.T) {
	for _, err := range []error{data.InvalidNickDataErr, data.NewerNickDataPresentErr, data.NickConflictErr} {
		// given
		repo, h, rr := makeComponents(t)

		buf := bytes.NewBuffer(makeJsonNickData(t))

		repo.putErr = err

		req, err := http.NewRequest("PUT", "/nicks", buf)
		if err != nil {
			t.Fatal(err)
		}

		// when
		h.ServeHTTP(rr, req)

		// then
		require.Equal(t, 400, rr.Code, "http status should be Bad Request for %s", err)
	}
}

func TestPutErr(t *testing.T) {
	// given
	repo, h, rr := makeComponents(t)

	buf := bytes.NewBuffer(makeJsonNickData(t))

	repo.putErr = errors.New("some error")

	req, err := http.NewRequest("PUT", "/nicks", buf)
	if err != nil {
		t.Fatal(err)
	}

	// when
	h.ServeHTTP(rr, req)

	// then
	require.Equal(t, 500, rr.Code, "http status should be Internal Server Error")
}

func TestList(t *testing.T) {
	// given
	repo, h, rr := makeComponents(t)

	repo.listReturn = make([]data.NickData, 0)

	req, err := http.NewRequest("GET", "/nicks", nil)
	if err != nil {
		t.Fatal(err)
	}

	// when
	h.ServeHTTP(rr, req)

	// then
	require.Equal(t, 200, rr.Code, "http status should be OK")
	require.Equal(t, "[]", rr.Body.String(), "body should contain an empty json array")
}

func TestListOneElement(t *testing.T) {
	// given
	repo, h, rr := makeComponents(t)

	repo.listReturn = []data.NickData{*makeNickData()}

	req, err := http.NewRequest("GET", "/nicks", nil)
	if err != nil {
		t.Fatal(err)
	}

	// when
	h.ServeHTTP(rr, req)

	// then
	expectedBody := `[{"id":"6964","nick":"nick","time":"1990-01-01T01:01:01.000000001Z","publicKey":"cHVibGljIGtleQ==","signature":"c2lnbmF0dXJl"}]`
	require.Equal(t, 200, rr.Code, "http status should be OK")
	require.Equal(t, expectedBody, rr.Body.String(), "body should contain a json array with one nick data")
}
