package wtf

import (
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	logger "github.com/i11cn/go_logger"
)

type (
	logger_wrapper struct {
		logger *logger.Logger
	}

	wtf_server struct {
		midware_chain []Midware
		access_logger Logger
		logger        Logger
		vhost         map[string]Mux
		tpl           Template
		builder       Builder
		arg_builder   map[string]func(Context) (reflect.Value, error)
	}
)

func init() {
	log := logger.GetLogger("wtf")
	log.AddAppender(logger.NewSplittedFileAppender("[%T] %M", "wtf-access.log", 24*time.Hour))
	log.SetLevel(logger.ALL)

	log = logger.GetLogger("wtf-debug").SetName("WTF")
	log.AddAppender(logger.NewConsoleAppender("[%T] [%N-%L] %f@%F.%l: %M"))
	log.AddAppender(logger.NewSplittedFileAppender("[%T] [%N-%L] %f@%F.%l: %M", "wtf.log", 24*time.Hour))
	log.SetLevel(logger.ALL)
}

func access_logger() Logger {
	log := logger.GetLogger("wtf")
	log.SkipPC(4)
	return &logger_wrapper{log}
}

func debug_logger() Logger {
	log := logger.GetLogger("wtf-debug")
	log.SkipPC(4)
	return &logger_wrapper{log}
}

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

// 用默认的组件们创建Server，默认组件是指Logger使用了github.com/i11cn/go_logger，
// 支持vhost，Mux使用了WTF自己实现的Mux，Context使用了WTF自己实现的Context，
// Template也是WTF实现的。
func NewServer(logger ...Logger) Server {
	builder := DefaultBuilder()
	ret := &wtf_server{}
	if len(logger) > 0 {
		ret.access_logger = logger[0]
	} else {
		ret.access_logger = access_logger()
	}
	ret.logger = debug_logger()
	ret.builder = builder

	ret.midware_chain = make([]Midware, 0, 10)
	ret.vhost = make(map[string]Mux)

	ret.arg_builder = make(map[string]func(Context) (reflect.Value, error))
	ret.arg_builder["wtf.Context"] = func(c Context) (reflect.Value, error) {
		return reflect.ValueOf(c), nil
	}
	ret.arg_builder["wtf.Request"] = func(c Context) (reflect.Value, error) {
		return reflect.ValueOf(c.Request()), nil
	}
	ret.arg_builder["*http.Request"] = func(c Context) (reflect.Value, error) {
		return reflect.ValueOf(c.HttpRequest()), nil
	}
	ret.arg_builder["wtf.Response"] = func(c Context) (reflect.Value, error) {
		return reflect.ValueOf(c.Response()), nil
	}
	ret.arg_builder["wtf.Rest"] = func(c Context) (reflect.Value, error) {
		return reflect.ValueOf(c.RestInfo()), nil
	}
	ret.arg_builder["http.ResponseWriter"] = func(c Context) (reflect.Value, error) {
		return reflect.ValueOf(c.HttpResponse()), nil
	}

	ret.tpl = NewTemplate()
	return ret
}

func (s *wtf_server) SetBuilder(b Builder) {
	s.builder = b
}

func (s *wtf_server) GetBuilder() Builder {
	return s.builder
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
	mux, exist := s.vhost[host]
	if !exist {
		mux = s.builder.BuildMux()
		s.vhost[host] = mux
	}
	if len(method) > 0 {
		return mux.Handle(h, p, method...)
	} else {
		return mux.Handle(h, p)
	}
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

func (s *wtf_server) ArgBuilder(fn interface{}) error {
	t := reflect.TypeOf(fn)
	if t.Kind() != reflect.Func {
		return fmt.Errorf("参数生成器只能接收函数类型，不能接收 %s 类型", t.String())
	}
	if t.NumIn() != 1 || t.In(0).String() != "wtf.Context" {
		return fmt.Errorf("入参必须是唯一的，且类型为 wtf.Context")
	}
	if t.NumOut() != 2 || t.Out(1).String() != "error" {
		return fmt.Errorf("必须有两个出参，且第二个必须为 error 类型")
	}
	typ := t.Out(0).String()
	v := reflect.ValueOf(fn)
	f := func(ctx Context) (reflect.Value, error) {
		ret := v.Call([]reflect.Value{reflect.ValueOf(ctx)})
		err, _ := ret[1].Interface().(error)
		return ret[0], err
	}
	s.arg_builder[typ] = f
	return nil
}

func (s *wtf_server) Handle(f interface{}, p string, args ...string) Error {
	if h, ok := f.(func(Context)); ok {
		return s.handle_func(h, p, args...)
	}
	t := reflect.TypeOf(f)
	if t.Kind() != reflect.Func {
		return NewError(0, "Handle的第一个参数必须是函数")
	}
	num_in := t.NumIn()
	if n, ok := f.(func(Context)); ok {
		return s.handle_func(n, p, args...)
	}
	v := reflect.ValueOf(f)
	a := make([]func(Context) (reflect.Value, error), 0, num_in)
	for i := 0; i < num_in; i++ {
		ts := t.In(i).String()
		if fn, exist := s.arg_builder[ts]; exist {
			a = append(a, fn)
		} else {
			return NewError(500, "不支持的参数类型: "+ts)
		}
	}
	h := func(c Context) {
		args := make([]reflect.Value, 0, num_in)
		for _, af := range a {
			v, e := af(c)
			if e != nil {
				s.logger.Error(e.Error())
				return
			}
			args = append(args, v)
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
	writer := s.builder.BuildWriter(s.logger, resp)
	ctx := s.builder.BuildContext(s.logger, req, writer, s.tpl, s.builder)
	defer func(c Context, w WriterWrapper) {
		w.Flush()
		info := w.GetWriteInfo()
		s.access_logger.Logf("[%d] [%d] %s %s", info.RespCode(), info.WriteBytes(), req.Method, req.URL.RequestURI())
	}(ctx, writer)

	host := req.URL.Hostname()
	mux, exist := s.vhost[strings.ToUpper(host)]
	if !exist {
		mux, exist = s.vhost["*"]
	}
	if !exist {
		ctx.Response().StatusCode(http.StatusInternalServerError, fmt.Sprintf("Unknow host name %s", host))
		s.logger.Errorf("未知Host: %s", host)
		return
	}
	for _, m := range s.midware_chain {
		ctx = m.Proc(ctx)
		if ctx == nil {
			return
		}
		defer func(w WriterWrapper) {
			w.Flush()
		}(ctx.HttpResponse())
	}
	handler, up := mux.Match(req)
	ctx.SetRestInfo(up)
	if handler != nil {
		handler(ctx)
	} else {
		ctx.Response().StatusCode(http.StatusNotFound)
	}
}
