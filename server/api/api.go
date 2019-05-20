// Package api implements a framework for creating a JSON API.
package api

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/boreq/starlight-nick-server/logging"
	"github.com/julienschmidt/httprouter"
)

var log = logging.New("api")

var InternalServerError = NewError(500, "Internal server error.")
var BadRequest = NewError(400, "Bad request.")
var NotFound = NewError(404, "Not found.")
var NotImplemented = NewError(501, "Not implemented.")

type Error interface {
	GetCode() int
	Error() string
	WithMessage(message string) Error
}

func NewError(code int, message string) Error {
	return apiError{code, message}
}

type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err apiError) GetCode() int {
	return err.Code
}

func (err apiError) Error() string {
	return err.Message
}

func (err apiError) WithMessage(message string) Error {
	return apiError{Code: err.Code, Message: message}
}

type Handle func(r *http.Request, p httprouter.Params) (interface{}, Error)

func Call(w http.ResponseWriter, r *http.Request, p httprouter.Params, handle Handle) error {
	code := 200
	response, apiErr := handle(r, p)
	if apiErr != nil {
		response = apiError{apiErr.GetCode(), apiErr.Error()}
		code = apiErr.GetCode()
	}
	j, err := json.Marshal(response)
	if err != nil {
		log.Error("marshal error", "err", err)
		j, _ = json.Marshal(InternalServerError)
		code = InternalServerError.GetCode()
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_, err = bytes.NewBuffer(j).WriteTo(w)
	return err
}

func Wrap(handle Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		Call(w, r, p, handle)
	}
}
