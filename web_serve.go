package web

import (
	. "net/http"
)

type (
	mid_chain_item struct {
		name string
		mid MiddleWare
	}
)
func (s *WebServe) init() {
}

func (s *WebServe) ServeHTTP(w ResponseWriter, r *Request) {
	f, p := s.router.Match(r.URL.Path, r.Method)
	c := &Context{w: w, r: r, params: p, serve: s, index: 0, chain_proc: true}
	if f == nil {
		c.proc = s.p404
	} else {
		c.mid_chain = make([]MiddleWare, len(s.mid_chain))
		for i, item := range s.mid_chain {
			c.mid_chain[i] = item.mid
		}
		c.proc = f
	}
		c.Next()
}