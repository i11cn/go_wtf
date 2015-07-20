package web

import (
	. "net/http"
)

func (s *WebServe) init() {
}

func (s *WebServe) ServeHTTP(w ResponseWriter, r *Request) {
	f, p := s.router.Match(r.URL.Path, r.Method)
	c := &Context{w: w, r: r, p: p}
	if f == nil {
		s.p404(c)
	} else {
		f(c)
	}
}