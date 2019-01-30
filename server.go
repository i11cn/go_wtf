package wtf

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"
	"github.com/i11cn/go_logger"
)

type (
	logger_wrapper struct {
		logger *logger.Logger
	}

	wtf_server struct {
		midware_chain []Midware
		logger        Logger
		vhost         map[string]Mux
		tpl           Template

		mux_builder func() Mux
		ctx_builder func(Logger, http.ResponseWriter, *http.Request, Template) Context
	}
)

func (l *logger_wrapper) Trace(arg ...interface{}) {
	l.logger.Trace(arg...)
}

func (l *logger_wrapper) Tracef(layout string, arg ...interface{}) {
	l.logger.Trace(fmt.Sprintf(layout, arg...))
}

func (l *logger_wrapper) Debug(arg ...interface{}) {
	l.logger.Debug(arg...)
}

func (l *logger_wrapper) Debugf(layout string, arg ...interface{}) {
	l.logger.Debug(fmt.Sprintf(layout, arg...))
}

func (l *logger_wrapper) Info(arg ...interface{}) {
	l.logger.Info(arg...)
}

func (l *logger_wrapper) Infof(layout string, arg ...interface{}) {
	l.logger.Info(fmt.Sprintf(layout, arg...))
}

func (l *logger_wrapper) Log(arg ...interface{}) {
	l.logger.Log(arg...)
}

func (l *logger_wrapper) Logf(layout string, arg ...interface{}) {
	l.logger.Log(fmt.Sprintf(layout, arg...))
}

func (l *logger_wrapper) Warn(arg ...interface{}) {
	l.logger.Warning(arg...)
}

func (l *logger_wrapper) Warnf(layout string, arg ...interface{}) {
	l.logger.Warning(fmt.Sprintf(layout, arg...))
}

func (l *logger_wrapper) Error(arg ...interface{}) {
	l.logger.Error(arg...)
}

func (l *logger_wrapper) Errorf(layout string, arg ...interface{}) {
	l.logger.Error(fmt.Sprintf(layout, arg...))
}

func (l *logger_wrapper) Fatal(arg ...interface{}) {
	l.logger.Fatal(arg...)
}

func (l *logger_wrapper) Fatalf(layout string, arg ...interface{}) {
	l.logger.Fatal(fmt.Sprintf(layout, arg...))
}

func init() {
	log := logger.GetLogger("wtf")
	log.AddAppender(logger.NewSplittedFileAppender("[%T] [%N-%L] : %M", "wtf.log", 24*time.Hour))
	log.SetLevel(logger.ALL)

	log = logger.GetLogger("wtf-debug")
	log.AddAppender(logger.NewSplittedFileAppender("[%T] [%N-%L] %f@%F.%l: %M", "wtf.log", 24*time.Hour))
	log.SetLevel(logger.ALL)
}

// 用默认的组件们创建Server，默认组件是指Logger使用了github.com/i11cn/go_logger，
// 支持vhost，Mux使用了WTF自己实现的Mux，Context使用了WTF自己实现的Context，
// Template也是WTF实现的。
func NewServer() Server {
	ret := &wtf_server{}
	ret.midware_chain = make([]Midware, 0, 10)
	ret.logger = &logger_wrapper{logger.GetLogger("wtf")}
	ret.vhost = make(map[string]Mux)
	ret.mux_builder = func() Mux {
		return NewWTFMux()
	}
	ret.ctx_builder = func(l Logger, resp http.ResponseWriter, req *http.Request, tpl Template) Context {
		return new_context(l, resp, req, tpl)
	}
	ret.tpl = NewTemplate()
	return ret
}

func (s *wtf_server) SetMuxBuilder(f func() Mux) {
	s.mux_builder = f
}

func (s *wtf_server) SetContextBuilder(f func(Logger, http.ResponseWriter, *http.Request, Template) Context) {
	s.ctx_builder = f
}

func (s *wtf_server) SetLogger(logger Logger) {
	s.logger = logger
}

func (s *wtf_server) SetTemplate(tpl Template) {
	s.tpl = tpl
}

func (s *wtf_server) Template() Template {
	return s.tpl
}

func (s *wtf_server) SetMux(mux Mux, vhosts ...string) {
	for _, h := range vhosts {
		s.vhost[h] = mux
	}
}

func (s *wtf_server) set_vhost_handle(h func(Context), p string, method []string, host string) Error {
	if mux, exist := s.vhost[host]; exist {
		if len(method) > 0 {
			return mux.Handle(h, p, method...)
		}
		return mux.Handle(h, p)
	}
	mux := s.mux_builder()
	var err Error
	if len(method) > 0 {
		err = mux.Handle(h, p, method...)
	} else {
		err = mux.Handle(h, p)
	}
	s.vhost[host] = mux
	return err
}

func (s *wtf_server) handle_func(f func(Context), p string, args ...string) Error {
	methods := []string{}
	all_methods := false
	vhosts := []string{}
	all_vhosts := false

	if len(args) > 0 {
		for _, arg := range args {
			arg = strings.ToUpper(arg)
			switch arg {
			case "ALL":
				methods = AllSupportMethods()
				all_methods = true

			case "*":
				vhosts = []string{"*"}
				all_vhosts = true

			default:
				if ValidMethod(arg) {
					if !all_methods {
						methods = append(methods, arg)
					}
				} else {
					if !all_vhosts {
						vhosts = append(vhosts, arg)
					}
				}
			}
		}
		if len(vhosts) == 0 {
			vhosts = []string{"*"}
		}
	} else {
		methods = AllSupportMethods()
		vhosts = []string{"*"}
	}
	for _, host := range vhosts {
		s.set_vhost_handle(f, p, methods, host)
	}
	return nil
}

func (s *wtf_server) Handle(f interface{}, p string, args ...string) Error {
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return NewError(0, "Handle的第一个参数必须是函数")
	}
	num_in := t.NumIn()
	if n, ok := f.(func(Context)); ok {
		return s.handle_func(n, p, args...)
	}
	v := reflect.ValueOf(f)
	a := make([]func(Context) reflect.Value, 0, num_in)
	for i := 0; i < num_in; i++ {
		switch ts := t.In(i).String(); ts {
		case "wtf.Context":
			a = append(a, func(c Context) reflect.Value {
				return reflect.ValueOf(c)
			})
		case "*http.Request":
			a = append(a, func(c Context) reflect.Value {
				return reflect.ValueOf(c.Request())
			})
		case "wtf.Response":
			a = append(a, func(c Context) reflect.Value {
				return reflect.ValueOf(NewResponse(c))
			})
		default:
			return NewError(0, "不支持的参数类型:" + ts)
		}
	}
	h := func(c Context) {
		args := make([]reflect.Value, 0, num_in)
		for _, af := range a {
			args = append(args, af(c))
		}
		v.Call(args)
	}
	return s.handle_func(h, p, args...)
}

func (s *wtf_server) validate_priority(p int, t reflect.Type) int {
	white_list := []string{"wtf.new_gzip_midware"}
	ts := t.String()
	for _, w := range white_list {
		if ts == w {
			return p
		}
	}
	if p < 0 {
		return 0
	}
	if p > 100 {
		return 100
	}
	return p
}

func (s *wtf_server) AddMidware(m Midware) {
	s.midware_chain = append(s.midware_chain, m)
	sort.SliceStable(s.midware_chain, func(i, j int) bool {
		pi := s.validate_priority(s.midware_chain[i].Priority(), reflect.TypeOf(s.midware_chain[i]))
		pj := s.validate_priority(s.midware_chain[j].Priority(), reflect.TypeOf(s.midware_chain[j]))
		return pi < pj
	})
}

func (s *wtf_server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	host := strings.ToUpper(req.URL.Hostname())
	ctx := s.ctx_builder(s.logger, resp, req, s.tpl)
	defer func(c Context) {
		info := c.GetContextInfo()
		s.logger.Logf("[%d] [%d] %s %s", info.RespCode(), info.WriteBytes(), req.Method, req.URL.RequestURI())
	}(ctx)
	mux, exist := s.vhost[host]
	if !exist {
		mux, exist = s.vhost["*"]
	}
	if !exist {
		ctx.WriteHeader(500)
		ctx.WriteString(fmt.Sprintf("Unknow host name %s", host))
		return
	}
	for _, m := range s.midware_chain {
		ctx = m.Proc(ctx)
		if ctx == nil {
			return
		}
		if flush, ok := ctx.(Flushable); ok {
			defer flush.Flush()
		}
	}
	handler, up := mux.Match(req)
	ctx.SetRESTParams(up)
	if handler != nil {
		handler(ctx)
	}
}
