package wtf

import (
	"net/http"
)

type (
	wtf_writer_wrapper struct {
		resp        http.ResponseWriter
		code        int
		code_writed bool
		bytes       int64
	}
)

func NewWriterWrapper(log Logger, resp http.ResponseWriter) WriterWrapper {
	ret := &wtf_writer_wrapper{}
	ret.resp = resp
	ret.code = http.StatusOK
	ret.code_writed = false
	ret.bytes = 0
	return ret
}

func (ww *wtf_writer_wrapper) Header() http.Header {
	return ww.resp.Header()
}

func (ww *wtf_writer_wrapper) Write(in []byte) (int, error) {
	if !ww.code_writed && ww.bytes == 0 && ww.code != http.StatusOK {
		ww.resp.WriteHeader(ww.code)
		ww.code_writed = true
	}
	n, err := ww.resp.Write(in)
	ww.bytes += int64(n)
	return n, err
}

func (ww *wtf_writer_wrapper) WriteHeader(code int) {
	if ww.bytes == 0 {
		ww.code = code
	}
}

func (ww *wtf_writer_wrapper) GetWriteInfo() WriteInfo {
	return &wtf_write_info{ww.code, ww.bytes}
}

func (ww *wtf_writer_wrapper) Flush() error {
	if !ww.code_writed && ww.bytes == 0 && ww.code != http.StatusOK {
		ww.resp.WriteHeader(ww.code)
		ww.code_writed = true
	}
	return nil
}
