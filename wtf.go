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

	ContextInfo interface {
		RespCode() int
		WriteBytes() int
	}

	Context interface {
		Logger() Logger
		Request() *http.Request
		Template(name string) *template.Template
		Header() http.Header
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
		Handle(Handler, string, ...string)
		Match(*http.Request) []Handler
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
		Handle(Handler, string, ...string)
		HandleFunc(func(Context), string, ...string)
	}
)

func init() {
}
