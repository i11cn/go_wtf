package wtf

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
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
		req         Request
		resp        Response
		rest_params RESTParams
		tpl         Template
		rc          int
		rc_writed   bool
		data        *wtf_context_info
		buf         *bytes.Buffer
	}
)

func (wci *wtf_context_info) RespCode() int {
	return wci.resp_code
}

func (wci *wtf_context_info) WriteBytes() int64 {
	return wci.write_count
}

func NewContext(log Logger, req *http.Request, resp http.ResponseWriter, tpl Template, b Builder) Context {
	ret := &wtf_context{}
	ret.logger = log
	ret.hresp = resp
	ret.hreq = req
	ret.req = b.BuildRequest(log, req)
	ret.resp = b.BuildResponse(log, resp, tpl)
	ret.tpl = tpl
	ret.rc = 0
	ret.rc_writed = false
	ret.data = &wtf_context_info{}
	ret.data.resp_code = 200
	ret.data.write_count = 0
	ret.buf = new(bytes.Buffer)
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

func (wc *wtf_context) Header() http.Header {
	return wc.hresp.Header()
}

func (wc *wtf_context) SetRESTParams(rp RESTParams) {
	wc.rest_params = rp
}

func (wc *wtf_context) RESTParams() RESTParams {
	return wc.rest_params
}

func (wc *wtf_context) GetBody() ([]byte, Error) {
	ret, err := ioutil.ReadAll(wc.HttpRequest().Body)
	if err != nil {
		return nil, NewError(500, "读取Body失败", err)
	}
	return ret, nil
}

func (wc *wtf_context) GetJsonBody(obj interface{}) Error {
	d, err := wc.GetBody()
	if err != nil {
		return err
	}
	e := json.Unmarshal(d, obj)
	if e != nil {
		return NewError(500, "解析Json数据失败", e)
	}
	return nil
}

func (wc *wtf_context) WriteHeader(code int) {
	wc.rc = code
}

func (wc *wtf_context) Write(data []byte) (n int, err error) {
	return wc.buf.Write(data)
}

func (wc *wtf_context) WriteString(str string) (n int, err error) {
	return wc.Write([]byte(str))
}

func (wc *wtf_context) WriteStream(src io.Reader) (n int64, err error) {
	ret, err := io.Copy(wc.buf, src)
	return ret, err
}

func (wc *wtf_context) Flush() error {
	if !wc.rc_writed {
		if wc.rc == 0 {
			wc.rc = http.StatusOK
		}
		wc.hresp.WriteHeader(wc.rc)
		wc.data.resp_code = wc.rc
		wc.rc_writed = true
	}
	n, err := io.Copy(wc.hresp, wc.buf)
	if err == nil {
		wc.data.write_count += n
		wc.buf.Reset()
	}
	return err
}

func (wc *wtf_context) GetContextInfo() ContextInfo {
	return wc.data
}
