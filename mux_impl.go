package wtf

import (
	"errors"
	"regexp"
	"strings"
)

type (
	vhost_mux struct {
		mux map[string]func(*Context) func(*Context)
		def func(*Context)
	}
)

func NewVHostMux() *vhost_mux {
	return &vhost_mux{make(map[string]func(*Context) func(*Context)), nil}
}

func (vm *vhost_mux) Handle(host string, fn func(*Context)) error {
	vm.mux[strings.ToLower(host)] = func(*Context) func(*Context) {
		return fn
	}
	return nil
}

func (vm *vhost_mux) Default(fn func(*Context)) error {
	vm.def = fn
	return nil
}

func (vm *vhost_mux) HandleSubMux(host string, mux Mux) error {
	vm.mux[strings.ToLower(host)] = func(ctx *Context) func(*Context) {
		return mux.Match(ctx)
	}
	return nil
}

func (vm *vhost_mux) Match(ctx *Context) func(*Context) {
	if fn, exist := vm.mux[strings.ToLower(ctx.URL.Host)]; exist {
		return fn(ctx)
	} else {
		return vm.def
	}
}

type (
	method_mux struct {
		mux map[string]func(*Context) func(*Context)
		def func(*Context)
	}
)

func NewMethodMux() *method_mux {
	return &method_mux{make(map[string]func(*Context) func(*Context)), nil}
}

func (mm *method_mux) Handle(method string, fn func(*Context)) error {
	mm.mux[strings.ToUpper(method)] = func(*Context) func(*Context) {
		return fn
	}
	return nil
}

func (mm *method_mux) Default(fn func(*Context)) error {
	mm.def = fn
	return nil
}

func (mm *method_mux) HandleSubMux(method string, mux Mux) error {
	mm.mux[strings.ToUpper(method)] = func(ctx *Context) func(*Context) {
		return mux.Match(ctx)
	}
	return nil
}

func (mm *method_mux) Match(ctx *Context) func(*Context) {
	if fn, exist := mm.mux[strings.ToUpper(ctx.Method)]; exist {
		return fn(ctx)
	} else {
		return mm.def
	}
}

type (
	mux_entry struct {
		pattern string
		regex   *regexp.Regexp
		entry   func(*Context) func(*Context)
	}

	regex_mux struct {
		mux []mux_entry
		def func(*Context)
	}
)

func NewRegexMux() *regex_mux {
	return &regex_mux{make([]mux_entry, 0, 10), nil}
}

func (rm *regex_mux) add_handle(pattern string, fn func(*Context) func(*Context)) error {
	if len(pattern) < 1 {
		return errors.New("正则表达式的模式长度太短")
	}
	if !strings.HasPrefix(pattern, "/") {
		pattern = strings.Join([]string{"/", pattern}, "")
	}
	if !(strings.HasPrefix(pattern, "^") || strings.HasPrefix(pattern, "\\A")) {
		pattern = strings.Join([]string{"^", pattern}, "")
	}
	if !strings.HasSuffix(pattern, "/") {
		pattern = strings.Join([]string{pattern, "/?"}, "")
	}
	if !(strings.HasSuffix(pattern, "$") || strings.HasPrefix(pattern, "\\z")) {
		pattern = strings.Join([]string{pattern, "$"}, "")
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return err
	}
	rm.mux = append(rm.mux, mux_entry{pattern, re, fn})
	return nil
}

func (rm *regex_mux) Handle(pattern string, fn func(*Context)) error {
	return rm.add_handle(pattern, func(*Context) func(*Context) {
		return fn
	})
}

func (rm *regex_mux) Default(fn func(*Context)) error {
	rm.def = fn
	return nil
}

func (rm *regex_mux) HandleSubMux(pattern string, mux Mux) error {
	return rm.add_handle(pattern, func(ctx *Context) func(*Context) {
		return mux.Match(ctx)
	})
}

type (
	RegexParams struct {
		subs  []string
		named map[string]string
	}
)

func (rp *RegexParams) Get(name string) string {
	if v, exist := rp.named[name]; exist {
		return v
	} else {
		return ""
	}
}

func (rp *RegexParams) GetByIndex(idx int) string {
	if idx >= 0 && idx < len(rp.subs) {
		return rp.subs[idx]
	} else {
		return ""
	}
}

func (rm *regex_mux) Match(ctx *Context) func(*Context) {
	uri := ctx.Uri
	for _, e := range rm.mux {
		if e.regex.MatchString(uri) {
			rp := &RegexParams{}
			subs := e.regex.FindStringSubmatch(uri)
			if len(subs) > 0 {
				rp.subs = subs
				names := e.regex.SubexpNames()
				if len(names) == len(subs) {
					for i := 0; i < len(names); i++ {
						if len(names[i]) > 0 {
							rp.named[names[i]] = subs[i]
						}
					}
				}
				ctx.RestParams = rp
			}
			return e.entry(ctx)
		}
	}
	return rm.def
}
