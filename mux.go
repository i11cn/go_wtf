package wtf

import (
	"fmt"
	"net/http"
	"strings"
)

type (
	handle_wrapper struct {
		proc func(Context)
	}

	mux_node interface {
		match(string) (bool, Handler)
		merge(string, Handler) bool
		dump(string)
	}

	base_node struct {
		pattern    string
		handler    Handler
		text_subs  []mux_node
		regex_subs []mux_node
		any_sub    *any_node
	}

	text_node struct {
		base_node
		p_len int
	}

	any_node struct {
		base_node
	}

	regex_node struct {
		base_node
	}

	wtf_mux struct {
		node mux_node
	}
)

func (bn *base_node) dump(prefix string) {
	path := fmt.Sprintf("%s%s", prefix, bn.pattern)
	if bn.handler != nil {
		fmt.Println(path)
	}
	for _, m := range bn.text_subs {
		m.dump(path)
	}
	for _, m := range bn.regex_subs {
		m.dump(path)
	}
	if bn.any_sub != nil {
		bn.any_sub.dump(path)
	}
}

func new_any_node(h Handler) *any_node {
	ret := &any_node{}
	ret.pattern = "*"
	ret.handler = h
	ret.text_subs = make([]mux_node, 0, 0)
	ret.regex_subs = make([]mux_node, 0, 0)
	return ret
}

func new_text_node(p string, h Handler) *text_node {
	ret := &text_node{}
	ret.pattern = p
	ret.p_len = len(p)
	ret.handler = h
	ret.text_subs = make([]mux_node, 0, 0)
	ret.regex_subs = make([]mux_node, 0, 0)
	return ret
}

func new_regex_node(p string, h Handler) *regex_node {
	ret := &regex_node{}
	ret.pattern = p
	ret.handler = h
	ret.text_subs = make([]mux_node, 0, 0)
	ret.regex_subs = make([]mux_node, 0, 0)
	return ret
}

func (an *any_node) match(path string) (bool, Handler) {
	return true, an.handler
}

func (an *any_node) merge(path string, h Handler) bool {
	return false
}

func (tn *text_node) match(path string) (bool, Handler) {
	fmt.Println("正在匹配路径： ", path)
	if tn.pattern[0] != path[0] {
		return false, nil
	}
	if strings.HasPrefix(path, tn.pattern) {
		p := path[tn.p_len:]
		if len(p) > 0 {
			for _, mux := range tn.text_subs {
				m, h := mux.match(p)
				if m {
					return true, h
				}
			}
			for _, mux := range tn.regex_subs {
				_, h := mux.match(p)
				if h != nil {
					return true, h
				}
			}
			if tn.any_sub != nil {
				return true, tn.any_sub.handler
			}
			return true, nil
		} else {
			return true, tn.handler
		}
	}
	return true, nil
}

func (tn *text_node) split(i int) {
	sub := new_text_node(tn.pattern[i:], tn.handler)
	sub.text_subs, sub.regex_subs, sub.any_sub = tn.text_subs, tn.regex_subs, tn.any_sub
	tn.pattern, tn.p_len, tn.handler, tn.text_subs, tn.regex_subs, tn.any_sub = tn.pattern[:i], i, nil, make([]mux_node, 0, 10), make([]mux_node, 0, 10), nil
	tn.text_subs = append(tn.text_subs, sub)
}

func (tn *text_node) merge(path string, h Handler) bool {
	p_len := len(path)
	s_len := len(tn.pattern)
	min_len := p_len
	if p_len > s_len {
		min_len = s_len
	}
	for i := 0; i < min_len; i++ {
		if tn.pattern[i] != path[i] {
			if i == 0 {
				return false
			} else {
				tn.split(i)
				sub := new_text_node(path[i:], h)
				tn.text_subs = append(tn.text_subs, sub)
			}
			return true
		}
	}
	if p_len == s_len {
		tn.handler = h
	} else if p_len > s_len {
		left := path[min_len:]
		for _, m := range tn.text_subs {
			if m.merge(left, h) {
				return true
			}
		}
		sub := new_text_node(left, h)
		tn.text_subs = append(tn.text_subs, sub)
	} else {
		tn.split(min_len)
		tn.handler = h
	}
	return true
}

func (rn *regex_node) match(path string) (bool, Handler) {
	return false, nil
}

func (rn *regex_node) merge(path string, h Handler) bool {
	return false
}

func (hw *handle_wrapper) Proc(c Context) {
	hw.proc(c)
}

func NewSimpleMux() *wtf_mux {
	ret := &wtf_mux{}
	return ret
}

func (sm *wtf_mux) Handle(h Handler, p string, args ...string) {
	if sm.node == nil {
		sm.node = new_text_node(p, h)
	} else {
		sm.node.merge(p, h)
	}
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
