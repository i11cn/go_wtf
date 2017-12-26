package wtf

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"io/ioutil"
	"net/http"
)

type (
	wtf_context_info struct {
		resp_code   int
		write_count int
	}

	wtf_context struct {
		logger      Logger
		resp        http.ResponseWriter
		req         *http.Request
		resp_code   ResponseCode
		rest_params RESTParams
		tpl         Template
		rc_setted   bool
		data        *wtf_context_info
	}
)

func (wci *wtf_context_info) RespCode() int {
	return wci.resp_code
}

func (wci *wtf_context_info) WriteBytes() int {
	return wci.write_count
}

func new_context(logger Logger, resp http.ResponseWriter, req *http.Request, rc ResponseCode, tpl Template) *wtf_context {
	ret := &wtf_context{}
	ret.logger = logger
	ret.resp = resp
	ret.req = req
	ret.resp_code = rc
	ret.tpl = tpl
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

func (wc *wtf_context) Response() Response {
	return new_response(wc, wc.resp_code)
}

func (wc *wtf_context) Execute(name string, obj interface{}) ([]byte, Error) {
	d, err := wc.tpl.Execute(name, obj)
	if err != nil {
		return nil, NewError(500, err.Error(), err)
	}
	return d, nil
}

func (wc *wtf_context) Header() http.Header {
	return wc.resp.Header()
}

func (wc *wtf_context) SetRESTParams(rp RESTParams) {
	wc.rest_params = rp
}

func (wc *wtf_context) RESTParams() RESTParams {
	return wc.rest_params
}

func (wc *wtf_context) GetBody() ([]byte, Error) {
	ret, err := ioutil.ReadAll(wc.Request().Body)
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
