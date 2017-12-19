package wtf

import (
	"fmt"
	"strings"
)

type (
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

func min(a, b int) int {
	if a > b {
		return b
	} else {
		return a
	}
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func (bn *base_node) deep_clone_from(src *base_node) {
	bn.pattern, bn.handler, bn.any_sub = src.pattern, src.handler, src.any_sub
	bn.text_subs = make([]mux_node, 0, max(len(src.text_subs), 10))
	bn.regex_subs = make([]mux_node, 0, max(len(src.regex_subs), 10))
	for _, s := range src.text_subs {
		bn.text_subs = append(bn.text_subs, s.deep_clone())
	}
	for _, s := range src.regex_subs {
		bn.regex_subs = append(bn.regex_subs, s.deep_clone())
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

func (an *any_node) deep_clone() mux_node {
	ret := &any_node{}
	ret.deep_clone_from(&an.base_node)
	return ret
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

func (tn *text_node) deep_clone() mux_node {
	ret := &text_node{}
	ret.deep_clone_from(&tn.base_node)
	ret.p_len = tn.p_len
	return ret
}

func (rn *regex_node) match(path string) (bool, Handler) {
	return false, nil
}

func (rn *regex_node) merge(path string, h Handler) bool {
	return false
}

func (rn *regex_node) deep_clone() mux_node {
	ret := &regex_node{}
	ret.deep_clone_from(&rn.base_node)
	return ret
}
