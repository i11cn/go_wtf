package wtf

import (
	"github.com/i11cn/go_logger"
	"regexp"
	"strings"
)

type (
	UrlParams struct {
		Name  string
		Value string
	}

	Router interface {
		AddEntry(pattern string, method string, entry func(*Context)) bool
		Match(url string, method string) (func(*Context), []UrlParams)
	}

	router_entry struct {
		pattern string
		regex   *regexp.Regexp
		entry   map[string]func(*Context)
	}

	default_router struct {
		router []router_entry
	}
)

func (r *default_router) AddEntry(pattern string, method string, entry func(*Context)) bool {
	if len(pattern) < 1 {
		return false
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
	for _, e := range r.router {
		if e.pattern == pattern {
			e.entry[method] = entry
			return true
		}
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return false
	}
	e := router_entry{pattern, re, map[string]func(*Context){}}
	ms := strings.Split(strings.ToUpper(method), ",")
	for _, m := range ms {
		e.entry[strings.Trim(m, " ")] = entry
	}
	r.router = append(r.router, e)
	return true
}

func (r *default_router) parse_url_params(re *regexp.Regexp, url string) []UrlParams {
	res := re.FindStringSubmatch(url)
	if len(res) <= 1 {
		return []UrlParams{}
	}
	ret := make([]UrlParams, len(res)-1)
	names := re.SubexpNames()
	if len(names) == len(res) {
		for i := 0; i < len(ret); i++ {
			ret[i].Name = names[i+1]
			ret[i].Value = res[i+1]
		}
	} else {
		for i := 0; i < len(ret); i++ {
			ret[i].Name = ""
			ret[i].Value = res[i+1]
		}
	}
	return ret
}

func (r *default_router) Match(url string, method string) (f func(*Context), up []UrlParams) {
	f = nil
	up = []UrlParams{}
	var exist bool
	for _, item := range r.router {
		if item.regex.MatchString(url) {
			log := logger.GetLogger("web")
			log.Trace("pattern : \"", item.pattern, "\", url : \"", url, "\"")
			if f, exist = item.entry[strings.ToUpper(method)]; exist {
				up = r.parse_url_params(item.regex, url)
			}
		}
	}
	return
}
