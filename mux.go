package wtf

import (
	"net/http"
	"strings"
)

type (
	mux_node interface {
		match_self(string, RESTParams) (bool, string, RESTParams)
		match(string, RESTParams) (bool, func(Context), RESTParams)
		merge(string, func(Context)) bool
		deep_clone() mux_node
	}

	wtf_mux struct {
		nodes map[string]mux_node
	}
)

func NewWTFMux() Mux {
	ret := &wtf_mux{make(map[string]mux_node)}
	return ret
}

func (sm *wtf_mux) handle_to_method(h func(Context), p string, method string) Error {
	if mux, exist := sm.nodes[method]; exist {
		tmp := mux.deep_clone()
		tmp.merge(p, h)
		sm.nodes[method] = tmp
	} else {
		sm.nodes[method] = parse_path(p, h)
	}
	return nil
}

func (sm *wtf_mux) Handle(h func(Context), p string, args ...string) Error {
	methods := []string{}
	if len(args) > 0 {
		for _, m := range args {
			switch strings.ToUpper(m) {
			case "ALL":
				methods = AllSupportMethods()
				break
			default:
				if ValidMethod(m) {
					methods = append(methods, strings.ToUpper(m))
				}
			}
		}
	} else {
		methods = AllSupportMethods()
	}
	for _, m := range methods {
		sm.handle_to_method(h, p, m)
	}
	return nil
}

func (sm *wtf_mux) Match(req *http.Request) (func(Context), RESTParams) {
	up := RESTParams{}
	method := strings.ToUpper(req.Method)
	if mux, exist := sm.nodes[method]; exist {
		_, h, up := mux.match(req.URL.Path, up)
		if h != nil {
			return h, up
		}
	}
	return nil, up
}
