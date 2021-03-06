package wtf

import (
	"net/http"
)

type (
	wtf_builder struct {
		writer func(Logger, http.ResponseWriter) WriterWrapper
		req    func(Logger, *http.Request) Request
		resp   func(Logger, http.ResponseWriter, Template) Response
		ctx    func(Logger, *http.Request, http.ResponseWriter, Template, Builder) Context
		mux    func() Mux
	}
)

func DefaultBuilder() Builder {
	ret := &wtf_builder{}
	ret.writer = NewWriterWrapper
	ret.req = NewRequest
	ret.resp = NewResponse
	ret.ctx = NewContext
	ret.mux = NewWTFMux
	return ret
}

func (b *wtf_builder) SetWriterBuilder(fn func(Logger, http.ResponseWriter) WriterWrapper) Builder {
	b.writer = fn
	return b
}

func (b *wtf_builder) SetRequestBuilder(fn func(Logger, *http.Request) Request) Builder {
	b.req = fn
	return b
}

func (b *wtf_builder) SetResponseBuilder(fn func(Logger, http.ResponseWriter, Template) Response) Builder {
	b.resp = fn
	return b
}

func (b *wtf_builder) SetContextBuilder(fn func(Logger, *http.Request, http.ResponseWriter, Template, Builder) Context) Builder {
	b.ctx = fn
	return b
}

func (b *wtf_builder) SetMuxBuilder(fn func() Mux) Builder {
	b.mux = fn
	return b
}

func (b *wtf_builder) BuildWriter(log Logger, resp http.ResponseWriter) WriterWrapper {
	return b.writer(log, resp)
}

func (b *wtf_builder) BuildRequest(log Logger, req *http.Request) Request {
	return b.req(log, req)
}

func (b *wtf_builder) BuildResponse(log Logger, resp http.ResponseWriter, tpl Template) Response {
	return b.resp(log, resp, tpl)
}

func (b *wtf_builder) BuildContext(log Logger, req *http.Request, resp http.ResponseWriter, tpl Template, builder Builder) Context {
	return b.ctx(log, req, resp, tpl, builder)
}

func (b *wtf_builder) BuildMux() Mux {
	return b.mux()
}
