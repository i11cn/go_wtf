package wtf

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

type (
	wtf_response struct {
		resp http.ResponseWriter
	}
)

func NewResponse(log Logger, resp http.ResponseWriter, tpl Template) Response {
	return &wtf_response{resp: resp}
}

func (resp *wtf_response) Header() http.Header {
	return resp.resp.Header()
}

func (resp *wtf_response) Write(in []byte) (int, error) {
	return resp.resp.Write(in)
}

func (resp *wtf_response) WriteHeader(code int) {
	resp.StatusCode(code)
}

func (resp *wtf_response) GetResponseInfo() ResponseInfo {
	return &wtf_context_info{}
}

func (resp *wtf_response) WriteString(s string) (int, error) {
	return io.WriteString(resp.resp, s)
}

func (resp *wtf_response) WriteStream(in io.Reader) (int64, error) {
	fmt.Println(resp)
	fmt.Println(resp.resp)
	fmt.Println(in)
	return io.Copy(resp.resp, in)
}

func (resp *wtf_response) WriteJson(obj interface{}) (int, error) {
	data, e := json.Marshal(obj)
	if e != nil {
		return 0, e
	}
	resp.Header().Set("Content-Type", "application/json;charset=utf-8")
	return resp.Write(data)
}

func (resp *wtf_response) WriteXml(obj interface{}) (int, error) {
	data, e := xml.Marshal(obj)
	if e != nil {
		return 0, e
	}
	resp.Header().Set("Content-Type", "application/xml;charset=utf-8")
	return resp.Write(data)
}
func (resp *wtf_response) SetHeader(key, value string) {
	resp.Header().Set(key, value)
}

func (resp *wtf_response) StatusCode(code int, body ...string) {
	resp.WriteHeader(code)
	if len(body) > 0 {
		resp.WriteString(body[0])
	}
}

func (resp *wtf_response) NotFound(body ...string) {
	resp.StatusCode(http.StatusNotFound, body...)
}

func (resp *wtf_response) Redirect(url string) {
	resp.SetHeader("Location", url)
	resp.StatusCode(http.StatusMovedPermanently)
}

func (resp *wtf_response) Follow(url string, body ...string) {
	resp.SetHeader("Location", url)
	resp.StatusCode(http.StatusSeeOther, body...)
}

func (resp *wtf_response) CrossOrigin(req Request, domain ...string) {
	if len(domain) > 0 {
		resp.SetHeader("Access-Control-Allow-Origin", domain[0])
		resp.SetHeader("Access-Control-Allow-Credentialls", "true")
		resp.SetHeader("Access-Control-Allow-Method", "GET, POST")
		return
	}
	origin := req.GetHeader("Origin")
	if len(origin) > 0 {
		resp.SetHeader("Access-Control-Allow-Origin", origin)
		resp.SetHeader("Access-Control-Allow-Credentialls", "true")
		resp.SetHeader("Access-Control-Allow-Method", "GET, POST")
	}
}
