package wtf

import (
	"github.com/i11cn/go_logger"
	. "net/http"
	"strings"
)

type (
	mid_chain_item struct {
		name string
		mid  MiddleWare
	}
)

func (s *WebServe) init() {
}

func (s *WebServe) ServeHTTP(w ResponseWriter, r *Request) {
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
