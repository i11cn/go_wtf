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
		match_self(string, RESTParams) (bool, Handler, string, RESTParams)
		match(string, RESTParams) (bool, Handler, RESTParams)
		merge(string, Handler) bool
		deep_clone() mux_node
	}
)

func (hw *handle_wrapper) Proc(c Context) {
	hw.proc(c)
}

func NewWTFMux() Mux {
	ret := &wtf_mux{make(map[string]mux_node)}
	return ret
}

func (sm *wtf_mux) handle_to_method(h Handler, p string, method string) Error {
	if mux, exist := sm.nodes[method]; exist {
		tmp := mux.deep_clone()
		tmp.merge(p, h)
		sm.nodes[method] = tmp
	} else {
		sm.nodes[method] = parse_path(p, h)
	}
	return nil
}

func (sm *wtf_mux) Handle(h Handler, p string, args ...string) Error {
	all_methods := []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "HEAD", "TRACE"}
	methods := []string{}
	if len(args) > 0 {
		for _, m := range args {
			switch strings.ToUpper(m) {
			case "ALL", "*", "":
				methods = all_methods
				break
			default:
				methods = append(methods, strings.ToUpper(m))
			}
		}
	} else {
		methods = all_methods
	}
	for _, m := range methods {
		sm.handle_to_method(h, p, m)
	}
	return nil
}

func (sm *wtf_mux) Match(req *http.Request) ([]Handler, RESTParams) {
	up := RESTParams{}
	method := strings.ToUpper(req.Method)
	if mux, exist := sm.nodes[method]; exist {
		_, h, up := mux.match(req.URL.Path, up)
		if h != nil {
			return []Handler{h}, up
		}
	}
	return []Handler{}, up
}
