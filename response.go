package wtf

import (
	"encoding/json"
	"encoding/xml"
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
