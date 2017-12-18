package wtf

import (
	"net/http"
	"strings"
)

type (
	handle_wrapper struct {
		proc func(Context)
	}

	mux_node interface {
		match(path string) Handler
	}

	text_node struct {
		pattern string
		p_len   int
		handler Handler
		subs    []mux_node
	}

	any_node struct {
		handler Handler
	}

	regex_node struct {
		pattern string
		handler Handler
		subs    []mux_node
	}

	simple_mux struct {
		router map[string]Handler
		test   Handler
		node   mux_node
	}
)

func (an *any_node) match(path string) Handler {
	return an.handler
}

func (tn *text_node) match(path string) Handler {
	if strings.HasPrefix(path, tn.pattern) {
		p := path[tn.p_len:]
		if len(p) > 0 {
			for _, mux := range tn.subs {
				h := mux.match(p)
				if h != nil {
					return h
				}
			}
			return nil
		} else {
			return tn.handler
		}
	}
	return nil
}

func (hw *handle_wrapper) Proc(c Context) {
	hw.proc(c)
}

func NewSimpleMux() *simple_mux {
	ret := &simple_mux{}
	ret.router = make(map[string]Handler)
	ret.test = nil
	return ret
}

func (sm *simple_mux) Handle(h Handler, p string, args ...string) {
	sm.test = h
}

func (sm *simple_mux) Match(req *http.Request) []Handler {
	if sm.test == nil {
		return []Handler{}
	} else {
		return []Handler{sm.test}
	}
}
