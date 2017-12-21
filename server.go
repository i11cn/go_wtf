package wtf

import (
	"fmt"
	"github.com/i11cn/go_logger"
	"net/http"
	"sort"
	"time"
)

type (
	logger_wrapper struct {
		logger *logger.Logger
	}

	chain_wrapper struct {
		priority int
		name     string
		chain    Chain
		pattern  string
	}

	wtf_server struct {
		chain_list []chain_wrapper
		logger     Logger
		vhost      map[string]Mux
		tpl        Template

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

func NewServer() Server {
	ret := &wtf_server{}
	ret.chain_list = make([]chain_wrapper, 0, 10)
	ret.logger = &logger_wrapper{logger.GetLogger("wtf")}
	ret.vhost = make(map[string]Mux)
	ret.mux_builder = func() Mux {
		return NewWTFMux()
	}
	ret.ctx_builder = func(l Logger, resp http.ResponseWriter, req *http.Request, tpl Template) Context {
		return new_context(l, resp, req, tpl)
	}
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

func (s *wtf_server) set_vhost_handle(h Handler, p string, method string, host string) Error {
	if mux, exist := s.vhost[host]; exist {
		if len(method) > 0 {
			return mux.Handle(h, p, method)
		} else {
			return mux.Handle(h, p)
		}
	} else {
		mux := s.mux_builder()
		var err Error
		if len(method) > 0 {
			err = mux.Handle(h, p, method)
		} else {
			err = mux.Handle(h, p)
		}
		s.vhost[host] = mux
		return err
	}
}

func (s *wtf_server) Handle(h Handler, p string, args ...string) Error {
	if len(args) > 0 {
		method := args[0]
		if len(args) > 1 {
			for _, vh := range args[1:] {
				err := s.set_vhost_handle(h, p, method, vh)
				if err != nil {
					return err
				}
			}
			return nil
		} else {
			return s.set_vhost_handle(h, p, method, "*")
		}
	} else {
		return s.set_vhost_handle(h, p, "", "*")
	}
}

func (s *wtf_server) HandleFunc(f func(Context), p string, args ...string) Error {
	return s.Handle(&handle_wrapper{f}, p, args...)
}

func (s *wtf_server) find_chain(name string) *chain_wrapper {
	for _, c := range s.chain_list {
		if c.name == name {
			return &c
		}
	}
	return nil
}

func (s *wtf_server) remove_chain(name string) *chain_wrapper {
	c := s.find_chain(name)
	if c != nil {
		old := s.chain_list
		s.chain_list = make([]chain_wrapper, 0, 10)
		for _, i := range old {
			if i.name != name {
				s.chain_list = append(s.chain_list, i)
			}
		}
	}
	return c
}

func (s *wtf_server) add_chain_item(h Chain, name string, priority int, pattern string, vals ...[]string) {
	s.remove_chain(name)
	s.chain_list = append(s.chain_list, chain_wrapper{priority, name, h, pattern})
	sort.SliceStable(s.chain_list, func(i, j int) bool {
		return s.chain_list[i].priority < s.chain_list[j].priority
	})
}

func (s *wtf_server) AddChain(h Chain, name string, priority int, pattern string, vals ...[]string) {
	switch {
	case priority < 10:
		priority = 10
	case priority >= 20 && priority < 30:
		priority = 30
	case priority >= 40:
		priority = 39
	}
	s.add_chain_item(h, name, priority, pattern, vals...)
}

func (s *wtf_server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	host := req.URL.Hostname()
	mux, exist := s.vhost[host]
	if !exist {
		mux, exist = s.vhost["*"]
	}
	ctx := s.ctx_builder(s.logger, resp, req, s.tpl)
	defer func(c Context) {
		info := c.GetContextInfo()
		s.logger.Logf("[%d] [%d]", info.RespCode(), info.WriteBytes())
	}(ctx)
	if !exist {
		ctx.WriteHeader(500)
		ctx.WriteString(fmt.Sprintf("Unknow host name %s", host))
		return
	}
	handlers, up := mux.Match(req)
	ctx.SetRESTParams(up)
	if len(handlers) == 0 {
		return
	}
	for _, h := range handlers {
		h.Proc(ctx)
	}
}
