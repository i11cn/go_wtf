package wtf

import (
	"regexp"
	"strings"
)

type (
	base_node struct {
		pattern    string
		handler    Handler
		text_subs  []mux_node
		regex_subs []mux_node
		other_subs []mux_node
		any_sub    *any_node
	}

	text_node struct {
		base_node
		p_len int
	}

	any_node struct {
		base_node
		name string
	}

	regex_node struct {
		base_node
		name string
		re   *regexp.Regexp
	}

	other_node struct {
		base_node
		name string
	}
)

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func parse_path(path string, h Handler) mux_node {
	if len(path) == 0 {
		return nil
	}
	switch path[0] {
	case '*':
		return new_any_node(h)
	case ':':
		pos := strings.Index(path, "/")
		if pos != -1 {
			left := path[pos:]
			name := path[:pos]
			ret := new_other_node(name, nil)
			if len(left) > 0 {
				ret.set_sub_node(parse_path(left, h))
			} else {
				ret.handler = h
			}
			return ret
		}
		return new_other_node(path, h)
	case '(':
		left := 0
		for i, c := range path {
			switch c {
			case '(':
				left++
			case ')':
				left--
			}
			if left == 0 {
				pattern := path[:i+1]
				left := path[i+1:]
				if len(left) > 0 {
					ret := new_regex_node(pattern, nil)
					ret.set_sub_node(parse_path(left, h))
					return ret
				}
				return new_regex_node(pattern, h)
			}
		}
		return nil
	default:
		pos := strings.IndexAny(path, "*:(")
		if pos != -1 {
			left := path[pos:]
			path = path[:pos]
			ret := new_text_node(path, nil)
			ret.set_sub_node(parse_path(left, h))
			return ret
		}
		return new_text_node(path, h)
	}
	return nil
}

func (bn *base_node) match_sub_nodes(path string, up RESTParams) (bool, Handler, RESTParams) {
	if path == "" {
		return true, bn.handler, up
	}
	for _, mux := range bn.text_subs {
		if m, h, rup := mux.match(path, up); m {
			if h != nil {
				return true, h, rup
			} else {
				break
			}
		}
	}
	for _, mux := range bn.regex_subs {
		if _, h, rup := mux.match(path, up); h != nil {
			return true, h, rup
		}
	}
	for _, mux := range bn.other_subs {
		if _, h, rup := mux.match(path, up); h != nil {
			return true, h, rup
		}
	}
	if bn.any_sub != nil {
		return true, bn.any_sub.handler, up
	}
	return true, nil, up

}

func (bn *base_node) merge_sub_node(path string, h Handler) {
	if len(path) > 0 {
		for _, s := range bn.text_subs {
			if s.merge(path, h) {
				return
			}
		}
		for _, s := range bn.regex_subs {
			if s.merge(path, h) {
				return
			}
		}
		for _, s := range bn.other_subs {
			if s.merge(path, h) {
				return
			}
		}
		bn.set_sub_node(parse_path(path, h))
	} else {
		bn.handler = h
	}
}

func (bn *base_node) set_sub_node(m mux_node) {
	if m == nil {
		return
	}
	switch m.(type) {
	case *text_node:
		bn.text_subs = append(bn.text_subs, m)
	case *regex_node:
		bn.regex_subs = append(bn.regex_subs, m)
	case *other_node:
		bn.other_subs = append(bn.other_subs, m)
	case *any_node:
		bn.any_sub = m.(*any_node)
	}
}

func (bn *base_node) deep_clone_from(src *base_node) {
	bn.pattern, bn.handler, bn.any_sub = src.pattern, src.handler, src.any_sub
	bn.text_subs = make([]mux_node, 0, max(len(src.text_subs), 10))
	bn.regex_subs = make([]mux_node, 0, max(len(src.regex_subs), 10))
	bn.other_subs = make([]mux_node, 0, max(len(src.other_subs), 10))
	for _, s := range src.text_subs {
		bn.text_subs = append(bn.text_subs, s.deep_clone())
	}
	for _, s := range src.regex_subs {
		bn.regex_subs = append(bn.regex_subs, s.deep_clone())
	}
	for _, s := range src.other_subs {
		bn.other_subs = append(bn.other_subs, s.deep_clone())
	}
	if src.any_sub != nil {
		node := src.any_sub.deep_clone()
		bn.any_sub = node.(*any_node)
	}
}

func new_any_node(h Handler) *any_node {
	ret := &any_node{}
	ret.pattern = "*"
	ret.handler = h
	ret.text_subs = make([]mux_node, 0, 0)
	ret.regex_subs = make([]mux_node, 0, 0)
	ret.other_subs = make([]mux_node, 0, 0)
	return ret
}

func new_text_node(p string, h Handler) *text_node {
	ret := &text_node{}
	ret.pattern = p
	ret.p_len = len(p)
	ret.handler = h
	ret.text_subs = make([]mux_node, 0, 0)
	ret.regex_subs = make([]mux_node, 0, 0)
	ret.other_subs = make([]mux_node, 0, 0)
	return ret
}

func new_regex_node(p string, h Handler, reg ...string) *regex_node {
	pattern := p
	if len(reg) > 0 {
		pattern = reg[0]
	}
	re, err := regexp.Compile("^" + pattern)
	if err != nil {
		return nil
	}
	ret := &regex_node{}
	ret.pattern = p
	ret.handler = h
	ret.re = re
	names := re.SubexpNames()
	if len(names) > 1 {
		ret.name = names[1]
	}
	ret.text_subs = make([]mux_node, 0, 0)
	ret.regex_subs = make([]mux_node, 0, 0)
	ret.other_subs = make([]mux_node, 0, 0)
	return ret
}

func new_other_node(p string, h Handler) *other_node {
	ret := &other_node{}
	ret.pattern = p
	ret.handler = h
	ret.name = p[1:]
	ret.text_subs = make([]mux_node, 0, 0)
	ret.regex_subs = make([]mux_node, 0, 0)
	ret.other_subs = make([]mux_node, 0, 0)
	return ret
}

func (an *any_node) match(path string, _ RESTParams) (bool, Handler, RESTParams) {
	return true, an.handler, RESTParams{}
}

func (an *any_node) match_self(path string, _ RESTParams) (bool, string, RESTParams) {
	return true, "", RESTParams{}
}

func (an *any_node) merge(path string, h Handler) bool {
	return false
}

func (an *any_node) deep_clone() mux_node {
	ret := &any_node{}
	ret.deep_clone_from(&an.base_node)
	return ret
}

func (tn *text_node) match(path string, up RESTParams) (bool, Handler, RESTParams) {
	if m, p, _ := tn.match_self(path, up); m {
		return tn.match_sub_nodes(p, up)
	} else {
		return false, nil, up
	}
}

func (tn *text_node) match_self(path string, up RESTParams) (bool, string, RESTParams) {
	if (tn.pattern[0] != path[0]) || (tn.p_len > len(path)) {
		return false, path, up
	}
	if strings.HasPrefix(path, tn.pattern) {
		return true, path[tn.p_len:], up
	}
	return false, path, up
}

func (tn *text_node) split(i int) {
	sub := new_text_node(tn.pattern[i:], tn.handler)
	sub.text_subs, sub.regex_subs, sub.any_sub = tn.text_subs, tn.regex_subs, tn.any_sub
	tn.pattern, tn.p_len, tn.handler, tn.text_subs, tn.regex_subs, tn.any_sub = tn.pattern[:i], i, nil, make([]mux_node, 0, 10), make([]mux_node, 0, 10), nil
	tn.text_subs = append(tn.text_subs, sub)
}

func (tn *text_node) merge(path string, h Handler) bool {
	p_len := len(path)
	s_len := tn.p_len
	min_len := min(p_len, s_len)
	pos := 0
	for ; pos < min_len; pos++ {
		if tn.pattern[pos] != path[pos] {
			break
		}
	}
	if pos == 0 {
		return false
	}
	if pos != tn.p_len {
		tn.split(pos)
	}
	left := path[pos:]
	tn.merge_sub_node(left, h)
	return true
}

func (tn *text_node) deep_clone() mux_node {
	ret := &text_node{}
	ret.deep_clone_from(&tn.base_node)
	ret.p_len = tn.p_len
	return ret
}

func (rn *regex_node) match(path string, up RESTParams) (bool, Handler, RESTParams) {
	if m, p, rup := rn.match_self(path, up); m {
		return rn.match_sub_nodes(p, rup)
	} else {
		return false, nil, rup
	}
}

func (rn *regex_node) match_self(path string, up RESTParams) (bool, string, RESTParams) {
	pos := rn.re.FindStringIndex(path)
	if pos != nil {
		return true, path[pos[1]:], up.Append(rn.name, path[:pos[1]])
	}
	return false, path, up
}

func (rn *regex_node) merge(path string, h Handler) bool {
	if strings.HasPrefix(path, rn.pattern) {
		p_len := len(path)
		s_len := len(rn.pattern)
		if p_len == s_len {
			rn.handler = h
		} else {
			rn.merge_sub_node(path[s_len:], h)
		}
		return true
	}
	return false
}

func (rn *regex_node) deep_clone() mux_node {
	ret := &regex_node{}
	ret.deep_clone_from(&rn.base_node)
	ret.re = rn.re
	ret.name = rn.name
	return ret
}

func (on *other_node) match_self(path string, up RESTParams) (bool, string, RESTParams) {
	pos := strings.Index(path, "/")
	if pos == -1 {
		return true, "", up.Append(on.name, path)
	}
	return true, path[pos:], up.Append(on.name, path[:pos])
}

func (on *other_node) match(path string, up RESTParams) (bool, Handler, RESTParams) {
	if m, p, rup := on.match_self(path, up); m {
		return on.match_sub_nodes(p, rup)
	} else {
		return false, nil, rup
	}
}

func (on *other_node) merge(path string, h Handler) bool {
	if strings.HasPrefix(path, on.pattern) {
		p_len := len(path)
		s_len := len(on.pattern)
		if p_len == s_len {
			on.handler = h
		} else {
			on.merge_sub_node(path[s_len:], h)
		}
		return true
	}
	return false
}

func (on *other_node) deep_clone() mux_node {
	ret := &other_node{}
	ret.deep_clone_from(&on.base_node)
	ret.name = on.name
	return ret
}
