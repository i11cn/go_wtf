package wtf

import (
	"bytes"
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

	WebService struct {
		path_config PathConfig
		router      Router
		fs_handler  http.Handler
		mid_chain   []mid_chain_item
		def_page    map[int]func(*Context)
	}
)

func NewWebService() *WebService {
	ret := &WebService{router: &default_router{}}
	ret.init()
	return ret
}

func (s *WebService) SetRouter(r Router) {
	s.router = r
}

func (s *WebService) GetRouter() Router {
	return s.router
}

func (s *WebService) Get(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "GET", entry)
}

func (s *WebService) Post(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "POST", entry)
}

func (s *WebService) Put(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "PUT", entry)
}

func (s *WebService) Delete(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "DELETE", entry)
}

func (s *WebService) Head(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "HEAD", entry)
}

func (s *WebService) Option(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "OPTION", entry)
}

func (s *WebService) Patch(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "PATCH", entry)
}

func (s *WebService) All(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "GET,POST,PUT,DELETE,HEAD,OPTION,PATCH", entry)
}

func (s *WebService) SetDefaultPage(code int, f func(*Context)) {
	s.def_page[code] = f
}

func (s *WebService) AppendMiddleWare(f MiddleWare, name string) {
	for _, i := range s.mid_chain {
		if i.name == name {
			i.mid = f
			return
		}
	}
	s.mid_chain = append(s.mid_chain, mid_chain_item{name, f})
}

func (s *WebService) DeleteMiddleWare(name string) {
	tmp := make([]mid_chain_item, 0, len(s.mid_chain))
	for _, i := range s.mid_chain {
		if i.name != name {
			tmp = append(tmp, i)
		}
	}
	s.mid_chain = tmp
}

func (s *WebService) init() {
	if len(s.path_config.HtDoc) < 1 {
		s.path_config.HtDoc = "./htdoc"
	}
	if len(s.path_config.TemplatePath) < 1 {
		s.path_config.HtDoc = "./template"
	}
	s.def_page = make(map[int]func(*Context))
	s.def_page[404] = func(c *Context) {
		c.WriteString("页面还在天上飞呢...是你在地上吹么？")
	}
	s.def_page[500] = func(c *Context) {
		c.WriteString("你干啥了？服务器都被你弄死了")
	}
	s.fs_handler = http.FileServer(http.Dir(s.path_config.HtDoc))
}

func (s *WebService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f, p := s.router.Match(r.URL.Path, r.Method)
	log := logger.GetLogger("web")
	log.Log(r.Method, " - ", r.URL.Path)
	c := &Context{w: w, r: r, params: p, service: s, tpl: &default_template{path: s.path_config.TemplatePath}}
	if f == nil {
		buf := bytes.NewBufferString(s.path_config.HtDoc)
		buf.WriteString(r.URL.Path)
		filename := buf.String()
		if file_exist(filename) {
			s.fs_handler.ServeHTTP(w, r)
		} else {
			c.WriteStatusCode(404)
		}
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
		c.Process()
	}
}
