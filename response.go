package wtf

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
)

type (
	wtf_response struct {
		ctx Context
	}
)

func new_response(ctx Context) Response {
	ret := &wtf_response{}
	ret.ctx = ctx
	return ret
}

func (resp *wtf_response) StatusCode(code int, body ...string) {
	resp.ctx.WriteHeader(code)
	if len(body) > 0 {
		resp.ctx.WriteString(body[0])
	}
}

func (resp *wtf_response) NotFound(body ...string) {
	resp.StatusCode(http.StatusNotFound, body...)
}

func (resp *wtf_response) Redirect(url string) {
	resp.ctx.Header().Set("Location", url)
	resp.StatusCode(http.StatusMovedPermanently)
}

func (resp *wtf_response) Follow(url string, body ...string) {
	resp.ctx.Header().Set("Location", url)
	resp.StatusCode(http.StatusSeeOther, body...)
}

func (resp *wtf_response) CrossOrigin(domain ...string) {
	if len(domain) > 0 {
		resp.ctx.Header().Set("Access-Control-Allow-Origin", domain[0])
		resp.ctx.Header().Set("Access-Control-Allow-Credentialls", "true")
		resp.ctx.Header().Set("Access-Control-Allow-Method", "GET, POST")
		return
	}
	origin := resp.ctx.Request().Header.Get("Origin")
	if len(origin) > 0 {
		resp.ctx.Header().Set("Access-Control-Allow-Origin", origin)
		resp.ctx.Header().Set("Access-Control-Allow-Credentialls", "true")
		resp.ctx.Header().Set("Access-Control-Allow-Method", "GET, POST")
	}
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
