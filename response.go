package wtf

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"io"
	"net/http"
)

type (
	wtf_write_info struct {
		resp_code   int
		write_count int64
	}

	wtf_response struct {
		resp        http.ResponseWriter
		tpl         Template
		code        int
		code_writed bool
		buf         *bytes.Buffer
	}
)

func (wci *wtf_write_info) RespCode() int {
	return wci.resp_code
}

func (wci *wtf_write_info) WriteBytes() int64 {
	return wci.write_count
}

func NewResponse(log Logger, resp http.ResponseWriter, tpl Template) Response {
	ret := &wtf_response{}
	ret.resp = resp
	ret.tpl = tpl
	return ret
}

func (resp *wtf_response) Header() http.Header {
	return resp.resp.Header()
}

func (resp *wtf_response) Write(in []byte) (int, error) {
	return resp.resp.Write(in)
}

func (resp *wtf_response) WriteHeader(code int) {
	resp.resp.WriteHeader(code)
}

func (resp *wtf_response) WriteString(s string) (int, error) {
	return io.WriteString(resp.resp, s)
}

func (resp *wtf_response) WriteStream(in io.Reader) (int64, error) {
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

func (resp *wtf_response) Execute(name string, obj interface{}) Error {
	d, err := resp.tpl.Execute(name, obj)
	if err != nil {
		return err
	}
	if _, err := resp.Write(d); err != nil {
		return NewError(http.StatusInternalServerError, "写入模板数据时发生错误", err)
	}
	return nil
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
	if origin != "" {
		resp.SetHeader("Access-Control-Allow-Origin", origin)
		resp.SetHeader("Access-Control-Allow-Credentialls", "true")
		resp.SetHeader("Access-Control-Allow-Method", "GET, POST")
	}
}
