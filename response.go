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
		Write(...interface{}) (int, error)
		WriteBytes(...[]byte) (int, error)
		WriteString(...string) (int, error)
		WriteHeader(int) Response
		RespCode() int
		Empty() bool
		Flush() (int, error)

		WriteJson(interface{}) error
		WriteXml(interface{}) error
		GetWriter() http.ResponseWriter
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

func (wr *WTFResponse) GetWriter() http.ResponseWriter {
	return wr.ResponseWriter
}

func (wr *WTFResponse) WriteHeader(code int) Response {
	if !wr.is_code_set {
		wr.resp_code = code
		wr.is_code_set = true
	}
	return wr
}

func (wr *WTFResponse) Write(objs ...interface{}) (int, error) {
	s := fmt.Sprint(objs...)
	return wr.buf.Write([]byte(s))
}

func (wr *WTFResponse) WriteBytes(datas ...[]byte) (int, error) {
	var buf bytes.Buffer
	for _, d := range datas {
		buf.Write(d)
	}
	return wr.buf.Write(buf.Bytes())
}

func (wr *WTFResponse) WriteString(strs ...string) (int, error) {
	var buf bytes.Buffer
	for _, s := range strs {
		buf.WriteString(s)
	}
	return wr.buf.Write(buf.Bytes())
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
	_, err = wr.WriteBytes(d)
	return err
}

func (wr *WTFResponse) WriteXml(obj interface{}) error {
	d, err := xml.Marshal(obj)
	if err != nil {
		return err
	}
	wr.Header().Set("Content-Type", "application/xml;charset=utf-8")
	_, err = wr.WriteBytes(d)
	return err
}
