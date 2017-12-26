package wtf

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

type (
	wtf_response struct {
		ctx      Context
		respCode ResponseCode
	}
)

func new_response(ctx Context, rc ResponseCode) Response {
	ret := &wtf_response{}
	ret.ctx = ctx
	ret.respCode = rc
	return ret
}

func (resp *wtf_response) StatusCode(code int, body ...string) {
	resp.respCode.StatusCode(resp.ctx, code, body...)
}

func (resp *wtf_response) NotFound(body ...string) {
	resp.respCode.StatusCode(resp.ctx, http.StatusNotFound, body...)
}

func (resp *wtf_response) Redirect(url string) {
	resp.ctx.Header().Set("Location", url)
	resp.ctx.WriteHeader(http.StatusMovedPermanently)
}

func (resp *wtf_response) Follow(url string, body ...string) {
	resp.ctx.Header().Set("Location", url)
	resp.ctx.WriteHeader(http.StatusSeeOther)
	if len(body) > 0 {
		resp.ctx.Write([]byte(body[0]))
	}
}

func (resp *wtf_response) Write(data []byte) (int, error) {
	return resp.ctx.Write(data)
}

func (resp *wtf_response) WriteString(str string) (int, error) {
	return resp.ctx.Write([]byte(str))
}

func (resp *wtf_response) WriteStream(stream io.Reader) (int, error) {
	return resp.ctx.WriteStream(stream)
}

func (resp *wtf_response) WriteJson(obj interface{}) (int, error) {
	data, e := json.Marshal(obj)
	if e != nil {
		return 0, e
	}
	resp.ctx.Header().Set("Content-Type", "application/json;charset=utf-8")
	return resp.ctx.Write(data)
}

func (resp *wtf_response) WriteXml(obj interface{}) (int, error) {
	data, e := xml.Marshal(obj)
	if e != nil {
		return 0, e
	}
	resp.ctx.Header().Set("Content-Type", "application/xml;charset=utf-8")
	return resp.ctx.Write(data)
}
