package wtf

import (
	"net/http"
)

type (
	Response interface {
		Header() http.Header
		Write([]byte) (int, error)
		WriteHeader(int)
		WriteJson(interface{}) error
		WriteXml(interface{})
	}

	WTFResponse struct {
		http.ResponseWriter
		resp_code   int
		is_code_set bool
	}
)

func NewResponse(resp http.ResponseWriter) *WTFResponse {
	return &WTFResponse{resp, 200, false}
}

func (wr *WTFResponse) WriteHeader(code int) {
	if !wr.is_code_set {
		wr.ResponseWriter.WriteHeader(code)
		wr.is_code_set = true
	}
}
