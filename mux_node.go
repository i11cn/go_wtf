package wtf

import (
	"fmt"
	"regexp"
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
		name string
		re   *regexp.Regexp
	}

	wtf_mux struct {
		nodes map[string]mux_node
	}
)

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
			ret := new_regex_node(name, nil, fmt.Sprintf("(?P<%s>[\\w\\-\\~]+)", name[1:]))
			ret.set_sub_node(parse_path(left, h))
			return ret
		} else {
			return new_regex_node(path, h, fmt.Sprintf("(?P<%s>[\\w\\-\\~]+)", path[1:]))
		}
	case '(':
		left := 1
		for i, c := range path {
			switch c {
			case '(':
				left += 1
			case ')':
				left -= 1
			}
			if left == 0 {
				pattern := path[:i]
				left := path[i:]
				if len(left) == 0 {
					return new_regex_node(pattern, h)
				} else {
					ret := new_regex_node(pattern, nil)
					ret.set_sub_node(parse_path(left, h))
					return ret
				}
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
		} else {
			return new_text_node(path, h)
		}
	}
	return nil
}

func (bn *base_node) set_sub_node(m mux_node) {
	switch m.(type) {
	case *text_node:
		bn.text_subs = append(bn.text_subs, m)
	case *regex_node:
		bn.regex_subs = append(bn.regex_subs, m)
	case *any_node:
		bn.any_sub = m.(*any_node)
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
	return ret
}

func (an *any_node) match(path string, _ RESTParams) (bool, Handler, RESTParams) {
	return true, an.handler, RESTParams{}
}

func (an *any_node) match_self(path string, _ RESTParams) (bool, Handler, string, RESTParams) {
	return true, an.handler, "", RESTParams{}
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
	if m, h, p, _ := tn.match_self(path, up); m {
		if len(p) > 0 {
			for _, mux := range tn.text_subs {
				if m, h, up := mux.match(p, up); m {
					if h != nil {
						return true, h, up
					} else {
						break
					}
				}
			}
			for _, mux := range tn.regex_subs {
				if _, h, up := mux.match(p, up); h != nil {
					return true, h, up
				}
			}
			if tn.any_sub != nil {
				return true, tn.any_sub.handler, up
			}
			return true, nil, up
		} else {
			return true, h, up
		}
	} else {
		return false, nil, up
	}
	//if tn.pattern[0] != path[0] {
	//	return false, nil, up
	//}
	//if strings.HasPrefix(path, tn.pattern) {
	//	p := path[tn.p_len:]
	//	if len(p) > 0 {
	//		for _, mux := range tn.text_subs {
	//			m, h, rup := mux.match(p, up)
	//			if m {
	//				return true, h, rup
	//			}
	//		}
	//		for _, mux := range tn.regex_subs {
	//			_, h, rup := mux.match(p, up)
	//			if h != nil {
	//				return true, h, rup
	//			}
	//		}
	//		if tn.any_sub != nil {
	//			return true, tn.any_sub.handler, up
	//		}
	//		return true, nil, up
	//	} else {
	//		return true, tn.handler, up
	//	}
	//}
	//return true, nil, up
}

func (tn *text_node) match_self(path string, up RESTParams) (bool, Handler, string, RESTParams) {
	if (tn.pattern[0] != path[0]) || (tn.p_len > len(path)) {
		return false, nil, path, up
	}
	if strings.HasPrefix(path, tn.pattern) {
		return true, tn.handler, path[tn.p_len:], up
	}
	return false, nil, path, up
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
	for i := 0; i < min_len; i++ {
		if tn.pattern[i] != path[i] {
			if i == 0 {
				return false
			} else {
				tn.split(i)
				tn.set_sub_node(parse_path(path[i:], h))
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
		tn.set_sub_node(parse_path(left, h))
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

func (rn *regex_node) match(path string, up RESTParams) (bool, Handler, RESTParams) {
	if m, h, p, rup := rn.match_self(path, up); m {
		if len(p) > 0 {
			for _, mux := range rn.text_subs {
				if m, h, rup := mux.match(p, rup); m {
					if h != nil {
						return true, h, rup
					} else {
						break
					}
				}
			}
			for _, mux := range rn.regex_subs {
				if _, h, rup := mux.match(p, rup); h != nil {
					return true, h, rup
				}
			}
			if rn.any_sub != nil {
				return true, rn.any_sub.handler, rup
			}
			return true, nil, rup
		} else {
			return true, h, rup
		}
	} else {
		return false, nil, rup
	}

	//pos := rn.re.FindStringIndex(path)
	//if pos == nil {
	//	return false, nil, up
	//}
	//up = up.Append(rn.name, path[:pos[1]])
	//left := path[pos[1]:]
	//if len(left) == 0 {
	//	return true, rn.handler, up
	//}
	//for _, mux := range rn.text_subs {
	//	m, h, rup := mux.match(left, up)
	//	if m {
	//		return true, h, rup
	//	}
	//}
	//for _, mux := range rn.regex_subs {
	//	_, h, rup := mux.match(left, up)
	//	if h != nil {
	//		return true, h, rup
	//	}
	//}
	//return false, nil, up
}

func (rn *regex_node) match_self(path string, up RESTParams) (bool, Handler, string, RESTParams) {
	pos := rn.re.FindStringIndex(path)
	if pos != nil {
		return true, rn.handler, path[pos[1]:], up.Append(rn.name, path[:pos[1]])
	}
	return false, nil, path, up
}

func (rn *regex_node) merge(path string, h Handler) bool {
	p_len := len(path)
	s_len := len(rn.pattern)
	if p_len >= s_len && rn.pattern == path[:s_len] {
		if p_len == s_len {
			rn.handler = h
		} else {
			rn.set_sub_node(parse_path(path[s_len:], h))
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
