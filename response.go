package wtf

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/http"
)

type (
	Response interface {
		Header() http.Header
		Write([]byte) (int, error)
		WriteHeader(int)
		RespCode() int
		Empty() bool
		Flush() (int, error)

		WriteJson(interface{}) error
		WriteXml(interface{}) error
	}

	WTFResponse struct {
		http.ResponseWriter
		buf         bytes.Buffer
		resp_code   int
		is_code_set bool
	}
)

func NewResponse(resp http.ResponseWriter) Response {
	return &WTFResponse{resp, bytes.Buffer{}, http.StatusOK, false}
}

func (wr *WTFResponse) WriteHeader(code int) *WTFResponse {
	if !wr.is_code_set {
		wr.resp_code = code
		wr.is_code_set = true
	}
	return wr
}

func (wr *WTFResponse) Write(d []byte) (int, error) {
	return wr.buf.Write(d)
}

func (wr *WTFResponse) RespCode() int {
	return wr.resp_code
}

func (wr *WTFResponse) Empty() bool {
	return wr.buf.Len() == 0
}

func (wr *WTFResponse) Flush() (int, error) {
	len := wr.buf.Len()
	if wr.resp_code != http.StatusOK {
		wr.ResponseWriter.WriteHeader(wr.resp_code)
	}
	wr.ResponseWriter.Header().Set("Content-Length", fmt.Sprintf("%d", len))
	return wr.ResponseWriter.Write(wr.buf.Bytes())
}

func (wr *WTFResponse) WriteJson(obj interface{}) error {
	d, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	wr.Header().Set("Content-Type", "application/json;charset=utf-8")
	_, err = wr.Write(d)
	return err
}

func (wr *WTFResponse) WriteXml(obj interface{}) error {
	d, err := xml.Marshal(obj)
	if err != nil {
		return err
	}
	wr.Header().Set("Content-Type", "application/xml;charset=utf-8")
	_, err = wr.Write(d)
	return err
}
