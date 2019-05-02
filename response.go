package wtf

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

type (
	wtf_response struct {
		Context
		resp http.ResponseWriter
	}
)

func NewResponse(Context Context) Response {
	ret := &wtf_response{}
	ret.Context = Context
	return ret
}

func ResponseBuilder(log Logger, resp http.ResponseWriter, tpl Template) Response {
	return &wtf_response{resp: resp}
}

func (resp *wtf_response) GetResponseInfo() ResponseInfo {
	return &wtf_context_info{}
}

func (resp *wtf_response) WriteString(s string) (int, error) {
	return io.WriteString(resp.Context, s)
}

func (resp *wtf_response) WriteStream(in io.Reader) (int64, error) {
	return io.Copy(resp.Context, in)
}

func (resp *wtf_response) WriteJson(obj interface{}) (int, error) {
	data, e := json.Marshal(obj)
	if e != nil {
		return 0, e
	}
	resp.Context.Header().Set("Content-Type", "application/json;charset=utf-8")
	return resp.Context.Write(data)
}

func (resp *wtf_response) WriteXml(obj interface{}) (int, error) {
	data, e := xml.Marshal(obj)
	if e != nil {
		return 0, e
	}
	resp.Context.Header().Set("Content-Type", "application/xml;charset=utf-8")
	return resp.Context.Write(data)
}

func (resp *wtf_response) StatusCode(code int, body ...string) {
	resp.Context.WriteHeader(code)
	if len(body) > 0 {
		resp.Context.WriteString(body[0])
	}
}

func (resp *wtf_response) NotFound(body ...string) {
	resp.StatusCode(http.StatusNotFound, body...)
}

func (resp *wtf_response) Redirect(url string) {
	resp.Context.Header().Set("Location", url)
	resp.StatusCode(http.StatusMovedPermanently)
}

func (resp *wtf_response) Follow(url string, body ...string) {
	resp.Context.Header().Set("Location", url)
	resp.StatusCode(http.StatusSeeOther, body...)
}

func (resp *wtf_response) CrossOrigin(domain ...string) {
	if len(domain) > 0 {
		resp.Context.Header().Set("Access-Control-Allow-Origin", domain[0])
		resp.Context.Header().Set("Access-Control-Allow-Credentialls", "true")
		resp.Context.Header().Set("Access-Control-Allow-Method", "GET, POST")
		return
	}
	origin := resp.Context.HttpRequest().Header.Get("Origin")
	if len(origin) > 0 {
		resp.Context.Header().Set("Access-Control-Allow-Origin", origin)
		resp.Context.Header().Set("Access-Control-Allow-Credentialls", "true")
		resp.Context.Header().Set("Access-Control-Allow-Method", "GET, POST")
	}
}
