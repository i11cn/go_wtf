package wtf

import (
	"io"
	"net/http"
)

type (
	Error interface {
		Error() string
		Code() int
		Message() string
	}

	Logger interface {
		Trace(...interface{})
		Tracef(string, ...interface{})
		Debug(...interface{})
		Debugf(string, ...interface{})
		Info(...interface{})
		Infof(string, ...interface{})
		Log(...interface{})
		Logf(string, ...interface{})
		Warn(...interface{})
		Warnf(string, ...interface{})
		Error(...interface{})
		Errorf(string, ...interface{})
		Fatal(...interface{})
		Fatalf(string, ...interface{})
	}

	Template interface {
		BindPipe(string, interface{})
		LoadText(string)
		LoadFiles(...string)
		Execute(string, interface{}) ([]byte, error)
	}

	RESTParam struct {
		name  string
		value string
	}

	RESTParams []RESTParam

	ContextInfo interface {
		RespCode() int
		WriteBytes() int
	}

	Response interface {
		StatusCode(int, ...string)
		NotFound(...string)
		Redirect(string)
		Follow(string, ...string)
		Write([]byte) (int, error)
		WriteString(string) (int, error)
		WriteStream(io.Reader) (int, error)
		WriteJson(interface{}) (int, error)
		WriteXml(interface{}) (int, error)
	}

	Context interface {
		Logger() Logger
		Request() *http.Request
		Response() Response
		Execute(string, interface{}) ([]byte, Error)
		Header() http.Header
		SetRESTParams(RESTParams)
		RESTParams() RESTParams
		GetBody() ([]byte, Error)
		GetJsonBody(interface{}) Error
		WriteHeader(int)
		Write([]byte) (int, error)
		WriteString(string) (int, error)
		WriteStream(io.Reader) (int, error)
		GetContextInfo() ContextInfo
	}

	Handler interface {
		Proc(Context)
	}

	ResponseCode interface {
		Handle(int, func(Context))
		StatusCode(Context, int, ...string)
	}

	Mux interface {
		Handle(Handler, string, ...string) Error
		Match(*http.Request) ([]Handler, RESTParams)
	}

	Chain interface {
		Proc(Context) bool
	}

	MuxBuilder     func() Mux
	ContextBuilder func(Logger, http.ResponseWriter, *http.Request, ResponseCode, Template) Context

	Server interface {
		http.Handler
		SetMuxBuilder(func() Mux)
		SetContextBuilder(func(Logger, http.ResponseWriter, *http.Request, ResponseCode, Template) Context)
		SetLogger(Logger)
		SetTemplate(Template)
		Template() Template
		SetMux(Mux, ...string)
		Handle(Handler, string, ...string) Error
		HandleFunc(func(Context), string, ...string) Error
		HandleResponseCode(int, func(Context))
	}
)

func init() {
}

func (p RESTParams) Get(name string) string {
	for _, i := range []RESTParam(p) {
		if i.name == name {
			return i.value
		}
	}
	return ""
}

func (p RESTParams) GetIndex(i int) string {
	pa := []RESTParam(p)
	if len(pa) > i {
		return pa[i].value
	}
	return ""
}

func (p RESTParams) Append(name, value string) RESTParams {
	ret := []RESTParam(p)
	ret = append(ret, RESTParam{name, value})
	return RESTParams(ret)
}
