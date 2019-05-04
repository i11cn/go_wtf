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
		builder     Builder
		hreq        *http.Request
		hresp       WriterWrapper
		req         Request
		resp        Response
		rest_params Rest
		tpl         Template
	}
)

func NewContext(log Logger, req *http.Request, resp http.ResponseWriter, tpl Template, b Builder) Context {
	ret := &wtf_context{}
	ret.logger = log
	ret.builder = b
	ret.hresp = b.BuildWriter(log, resp)
	ret.hreq = req
	ret.req = b.BuildRequest(log, req)
	ret.resp = b.BuildResponse(log, resp, tpl)
	ret.tpl = tpl
	return ret
}

func (wc *wtf_context) Logger() Logger {
	return wc.logger
}

func (wc *wtf_context) Builder() Builder {
	return wc.builder
}

func (wc *wtf_context) HttpRequest() *http.Request {
	return wc.hreq
}

func (wc *wtf_context) Request() Request {
	return wc.req
}

func (wc *wtf_context) HttpResponse() WriterWrapper {
	return wc.hresp
}

func (wc *wtf_context) Response() Response {
	return wc.resp
}

func (wc *wtf_context) Template() Template {
	return wc.tpl
}

func (wc *wtf_context) SetRestInfo(rp Rest) {
	wc.rest_params = rp
}

func (wc *wtf_context) RestInfo() Rest {
	return wc.rest_params
}

func (wc *wtf_context) Clone(writer ...WriterWrapper) Context {
	ret := &wtf_context{
		logger:      wc.logger,
		builder:     wc.builder,
		hreq:        wc.hreq,
		hresp:       wc.hresp,
		req:         wc.req,
		resp:        wc.resp,
		rest_params: wc.rest_params,
		tpl:         wc.tpl,
	}
	if len(writer) > 0 {
		w := writer[0]
		ret.hresp = w
		ret.resp = wc.builder.BuildResponse(ret.logger, w, ret.tpl)
	}
	return ret
}
