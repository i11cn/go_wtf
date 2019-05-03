package wtf

import (
	"net/http"
)

type (
	wtf_context_info struct {
		resp_code   int
		write_count int64
	}

	wtf_context struct {
		logger      Logger
		hreq        *http.Request
		hresp       http.ResponseWriter
		writer      WriterWrapper
		req         Request
		resp        Response
		rest_params Rest
		tpl         Template
	}
)

func NewContext(log Logger, req *http.Request, resp http.ResponseWriter, tpl Template, b Builder) Context {
	ret := &wtf_context{}
	ret.logger = log
	ret.hresp = resp
	ret.hreq = req
	ret.req = b.BuildRequest(log, req)
	ret.resp = b.BuildResponse(log, resp, tpl)
	ret.tpl = tpl
	return ret
}

func (wc *wtf_context) Logger() Logger {
	return wc.logger
}

func (wc *wtf_context) HttpRequest() *http.Request {
	return wc.hreq
}

func (wc *wtf_context) Request() Request {
	return wc.req
}

func (wc *wtf_context) HttpResponse() http.ResponseWriter {
	return wc.hresp
}

func (wc *wtf_context) Response() Response {
	return wc.resp
}

func (wc *wtf_context) Execute(name string, obj interface{}) ([]byte, Error) {
	d, err := wc.tpl.Execute(name, obj)
	if err != nil {
		return nil, NewError(500, err.Error(), err)
	}
	return d, nil
}

func (wc *wtf_context) SetRestInfo(rp Rest) {
	wc.rest_params = rp
}

func (wc *wtf_context) RestInfo() Rest {
	return wc.rest_params
}
