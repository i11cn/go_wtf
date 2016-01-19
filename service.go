package wtf

import (
	"github.com/i11cn/go_logger"
	"net/http"
	"strings"
)

type (
	PathConfig struct {
		HtDoc        string
		TemplatePath string
		JsPath       string
		CSSPath      string
		ImagePath    string
	}

	UrlParams struct {
		Name  string
		Value string
	}

	MiddleWare func(c *Context) bool

	Router interface {
		AddEntry(pattern string, method string, entry func(*Context)) bool
		Handle(entries interface{}) bool
		Match(url string, method string) (func(*Context), []UrlParams)
	}

	WebServe struct {
		path_config PathConfig
		router      Router
		fs_handler  http.Handler
		mid_chain   []mid_chain_item
		p404        func(*Context)
		p500        func(*Context)
	}

	WebService struct {
		path_config PathConfig
		router      Router
		fs_handler  http.Handler
		mid_chain   []mid_chain_item
		def_page    map[int]func(*Context)
	}
)

func NewWebService(pc *PathConfig) *WebService {
	ret := &WebService{router: &default_router{}}
	if pc == nil {
		ret.path_config = PathConfig{HtDoc: "./htdoc", TemplatePath: "./template"}
	} else {
		ret.path_config = *pc
		if len(ret.path_config.HtDoc) < 1 {
			ret.path_config.HtDoc = "./htdoc"
		}
		if len(ret.path_config.TemplatePath) < 1 {
			ret.path_config.HtDoc = "./template"
		}
	}
	ret.def_page[404] = func(c *Context) {
		c.WriteString("页面还在天上飞呢...是你在地上吹么？")
	}
	ret.def_page[500] = func(c *Context) {
		c.WriteString("你干啥了？服务器都被你弄死了")
	}
	ret.fs_handler = http.FileServer(http.Dir(ret.path_config.HtDoc))
	ret.init()
	return ret
}

func NewWebServe(pc *PathConfig) *WebServe {
	ret := &WebServe{router: &default_router{}}
	if pc == nil {
		ret.path_config = PathConfig{HtDoc: "./htdoc", TemplatePath: "./template"}
	} else {
		ret.path_config = *pc
		if len(ret.path_config.HtDoc) < 1 {
			ret.path_config.HtDoc = "./htdoc"
		}
		if len(ret.path_config.TemplatePath) < 1 {
			ret.path_config.HtDoc = "./template"
		}
	}
	ret.p404 = func(c *Context) {
		c.WriteString("页面还在天上飞呢...是你在地上吹么？")
	}
	ret.p500 = func(c *Context) {
		c.WriteString("你干啥了？服务器都被你弄死了")
	}
	ret.fs_handler = http.FileServer(http.Dir(ret.path_config.HtDoc))
	return ret
}

func (s *WebServe) SetRouter(r Router) {
	s.router = r
}

func (s *WebServe) GetRouter() Router {
	return s.router
}

func (s *WebServe) Get(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "GET", entry)
}

func (s *WebServe) Post(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "POST", entry)
}

func (s *WebServe) Put(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "PUT", entry)
}

func (s *WebServe) Delete(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "DELETE", entry)
}

func (s *WebServe) Head(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "HEAD", entry)
}

func (s *WebServe) Option(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "OPTION", entry)
}

func (s *WebServe) Patch(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "PATCH", entry)
}

func (s *WebServe) All(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "GET,POST,PUT,DELETE,HEAD,OPTION,PATCH", entry)
}

func (s *WebServe) Handle(entries interface{}) bool {
	return s.router.Handle(entries)
}

func (s *WebService) SetDefaultPage(code int, f func(*Context)) {
	s.def_page[code] = f
}

func (s *WebServe) Set404Page(f func(*Context)) {
	s.p404 = f
}

func (s *WebServe) AppendMiddleWare(f MiddleWare, name string) {
	for _, i := range s.mid_chain {
		if i.name == name {
			i.mid = f
			return
		}
	}
	s.mid_chain = append(s.mid_chain, mid_chain_item{name, f})
}

func (s *WebServe) DeleteMiddleWare(name string) {
	tmp := make([]mid_chain_item, 0, len(s.mid_chain))
	for _, i := range s.mid_chain {
		if i.name != name {
			tmp = append(tmp, i)
		}
	}
	s.mid_chain = tmp
}

func (s *WebService) init() {
}

func (s *WebServe) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, p := s.router.Match(r.URL.Path, r.Method)
	log := logger.GetLogger("web")
	log.Log(r.Method, " - ", r.URL.Path)
	c := &Context{w: w, r: r, params: p, serve: s, tpl: &default_template{path: s.path_config.TemplatePath}, index: 0, chain_proc: true}
	if f == nil {
		s.fs_handler.ServeHTTP(w, r)
		//c.w.WriteHeader(404)
		//c.proc = s.p404
	} else {
		parts := strings.Split(r.URL.RawQuery, "&")
		c.querys = make(map[string]string)
		for _, p := range parts {
			kv := strings.SplitN(p, "=", 2)
			if len(kv) == 2 {
				c.querys[kv[0]] = kv[1]
			}
		}
		c.mid_chain = make([]MiddleWare, len(s.mid_chain))
		for i, item := range s.mid_chain {
			c.mid_chain[i] = item.mid
		}
		c.proc = f
		c.Next()
	}
}
