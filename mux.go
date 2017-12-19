package wtf

import (
	"fmt"
	"net/http"
)

type (
	handle_wrapper struct {
		proc func(Context)
	}

	mux_node interface {
		match(string) (bool, Handler)
		merge(string, Handler) bool
		deep_clone() mux_node
		dump(string)
	}
)

func (hw *handle_wrapper) Proc(c Context) {
	hw.proc(c)
}

func NewSimpleMux() *wtf_mux {
	ret := &wtf_mux{}
	return ret
}

func (sm *wtf_mux) Handle(h Handler, p string, args ...string) Error {
	if sm.node == nil {
		sm.node = new_text_node(p, h)
	} else {
		tmp := sm.node.deep_clone()
		tmp.merge(p, h)
		sm.node = tmp
		//sm.node.merge(p, h)
	}
	return nil
}

func (sm *wtf_mux) Match(req *http.Request) []Handler {
	if sm.node != nil {
		_, h := sm.node.match(req.URL.Path)
		if h != nil {
			fmt.Println("匹配成功")
			return []Handler{h}
		}
	}
	return []Handler{}
}
