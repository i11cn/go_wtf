package wtf

import (
	"errors"
	"regexp"
	"strings"
)

type (
	Mux interface {
		Handle(string, func(*Context)) error
		HandleSubMux(string, Mux) error
		Match(*Context) func(*Context)
	}

	MethodMux struct {
		mux map[string]func(*Context) func(*Context)
	}
)

func NewMethodMux() *MethodMux {
	return &MethodMux{make(map[string]func(*Context) func(*Context))}
}

func (mm *MethodMux) Handle(method string, fn func(*Context)) error {
	mm.mux[strings.ToUpper(method)] = func(*Context) func(*Context) {
		return fn
	}
	return nil
}

func (mm *MethodMux) HandleSubMux(method string, mux Mux) error {
	mm.mux[strings.ToUpper(method)] = func(ctx *Context) func(*Context) {
		return mux.Match(ctx)
	}
	return nil
}

func (mm *MethodMux) Match(ctx *Context) func(*Context) {
	if fn, exist := mm.mux[ctx.Request.Method]; exist {
		return fn(ctx)
	} else {
		return nil
	}
}

func (mm *MethodMux) Handler() func(*Context) {
	return func(ctx *Context) {
		fn := mm.Match(ctx)
		if fn != nil {
			fn(ctx)
		} else {
		}
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
	}
)

func NewRegexMux() *regex_mux {
	return &regex_mux{make([]mux_entry, 0, 10)}
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

func (rm *regex_mux) HandleSubMux(pattern string, mux Mux) error {
	return rm.add_handle(pattern, func(ctx *Context) func(*Context) {
		return mux.Match(ctx)
	})
}

func (rm *regex_mux) Match(ctx *Context) func(*Context) {
	uri := ctx.Uri
	for _, e := range rm.mux {
		if e.regex.MatchString(uri) {
			return e.entry(ctx)
		}
	}
	return nil
}
