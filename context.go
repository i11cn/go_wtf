package wtf

import (
	"encoding/json"
	"encoding/xml"
	"html/template"
	"io"
	"net/http"
)

type (
	wtf_context_info struct {
		resp_code   int
		write_count int
	}

	wtf_context struct {
		logger    Logger
		resp      http.ResponseWriter
		req       *http.Request
		tpl       *template.Template
		rc_setted bool
		data      *wtf_context_info
	}
)

func (wci *wtf_context_info) RespCode() int {
	return wci.resp_code
}

func (wci *wtf_context_info) WriteBytes() int {
	return wci.write_count
}

func new_context(logger Logger, resp http.ResponseWriter, req *http.Request) *wtf_context {
	ret := &wtf_context{}
	ret.logger = logger
	ret.resp = resp
	ret.req = req
	ret.rc_setted = false
	ret.data = &wtf_context_info{}
	ret.data.resp_code = 200
	ret.data.write_count = 0
	return ret
}

func (wc *wtf_context) Logger() Logger {
	return wc.logger
}

func (wc *wtf_context) Request() *http.Request {
	return wc.req
}

func (wc *wtf_context) Template(name string) *template.Template {
	return wc.tpl
}

func (wc *wtf_context) Header() http.Header {
	return wc.resp.Header()
}

func (wc *wtf_context) WriteHeader(code int) {
	if !wc.rc_setted {
		wc.data.resp_code = code
		wc.rc_setted = true
	}
	wc.resp.WriteHeader(code)
}

func (wc *wtf_context) Write(data []byte) (n int, err error) {
	n, err = wc.resp.Write(data)
	wc.data.write_count += n
	return
}

func (wc *wtf_context) WriteString(str string) (n int, err error) {
	n, err = wc.resp.Write([]byte(str))
	wc.data.write_count += n
	return
}

func (wc *wtf_context) WriteStream(src io.Reader) (n int, err error) {
	ret, err := io.Copy(wc.resp, src)
	n = int(ret)
	wc.data.write_count += n
	return
}

func (wc *wtf_context) WriteJson(obj interface{}) (n int, err error) {
	data, e := json.Marshal(obj)
	if e != nil {
		return 0, e
	}
	wc.resp.Header().Set("Content-Type", "application/json;charset=utf-8")
	n, err = wc.resp.Write(data)
	wc.data.write_count += n
	return
}

func (wc *wtf_context) WriteXml(obj interface{}) (n int, err error) {
	data, e := xml.Marshal(obj)
	if e != nil {
		return 0, e
	}
	wc.resp.Header().Set("Content-Type", "application/xml;charset=utf-8")
	n, err = wc.resp.Write(data)
	wc.data.write_count += n
	return
}

func (wc *wtf_context) GetContextInfo() ContextInfo {
	return wc.data
}
