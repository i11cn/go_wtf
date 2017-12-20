package wtf

import (
	"html/template"
	"io"
	"net/http"
)

type (
	Error interface {
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
		Load(string)
		Loads(...string)
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

	Context interface {
		Logger() Logger
		Request() *http.Request
		Template(name string) *template.Template
		Header() http.Header
		SetRESTParams(RESTParams)
		RESTParams() RESTParams
		WriteHeader(int)
		Write([]byte) (int, error)
		WriteString(string) (int, error)
		WriteStream(io.Reader) (int, error)
		WriteJson(interface{}) (int, error)
		WriteXml(interface{}) (int, error)
		GetContextInfo() ContextInfo
	}

	Handler interface {
		Proc(Context)
	}

	Mux interface {
		Handle(Handler, string, ...string) Error
		Match(*http.Request) ([]Handler, RESTParams)
	}

	Chain interface {
		Proc(Context) bool
	}

	ErrorPage interface {
		SetPage(int, func(Context))
		Proc(int, Context)
	}

	Server interface {
		http.Handler
		SetMuxBuilder(func() Mux)
		SetContextBuilder(func(Logger, http.ResponseWriter, *http.Request) Context)
		SetLogger(Logger)
		SetTemplate(Template)
		Template() Template
		SetMux(Mux, ...string)
		SetErrorPage(int, func(Context))
		SetErrorPages(ErrorPage)
		Handle(Handler, string, ...string) Error
		HandleFunc(func(Context), string, ...string) Error
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
	} else {
		return ""
	}
}

func (p RESTParams) Append(name, value string) RESTParams {
	ret := []RESTParam(p)
	ret = append(ret, RESTParam{name, value})
	return RESTParams(ret)
}
