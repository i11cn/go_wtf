package web

import (
	"encoding/json"
	"encoding/xml"
	. "net/http"
)

type PathConfig  struct {
	HtDoc string
	TemplatePath string
	JsPath string
	CSSPath string
	ImagePath string
}

type UrlParams struct {
	Name string
	Value string
}

type Context struct {
	w ResponseWriter
	r *Request
	p []UrlParams
	mid_chain []func(*Context)
}

func (c *Context) SetMime(mime string) {
	c.w.Header().Set("Content-Type", mime)
}

func (c *Context) GetParamByIndex(i int) string {
	if len(c.p) >= i {
		return c.p[i].Value
	}
	return ""
}

func (c *Context) GetParam(name string) string {
	if len(name) > 0 {
		for _, s := range c.p {
			if s.Name == name {
				return s.Value
			}
		}
	}
	return ""
}

func (c *Context) WriteStatusCode(s int) {
	c.w.WriteHeader(s)
}
func (c *Context) Write(d []byte) (int, error) {
	return c.w.Write(d)
}

func (c *Context) WriteString(s string) error {
	_, err := c.w.Write([]byte(s))
	return err
}

func (c *Context) WriteJson(obj interface{}) error {
	d, err := json.Marshal(obj)
	if err == nil {
		c.SetMime("application/json")
		_, err = c.Write(d)
	}
	return err
}

func (c *Context) WriteXml(obj interface{}) error {
	d, err := xml.Marshal(obj)
	if err == nil {
		c.SetMime("application/xml")
		_, err = c.Write(d)
	}
	return err
}

type Router interface {
	AddEntry(pattern string, method string, entry func(*Context)) bool
	Handle(entries interface{}) bool
	Match(url string, method string) (func(*Context), []UrlParams)
}

type WebServe struct {
	router Router
	p404 func(*Context)
	p500 func(*Context)
}

func NewWebServe() *WebServe {
	ret := &WebServe{router: &default_router{}}
	ret.p404 = func(c *Context) {
		c.WriteStatusCode(404)
		c.WriteString("页面还在天上飞呢...是你在地上吹么？")
	}
	ret.p500 = func(c *Context) {
		c.WriteStatusCode(500)
		c.WriteString("你干啥了？服务器都被你弄死了")
	}
	ret.init()
	return ret
}

func (s *WebServe) SetRouter(r Router) {
	s.router = r
}

func (s *WebServe) Get(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "GET", entry)
}

func (s *WebServe) Post(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "POST", entry)
}

func (s *WebServe) Put(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "PUT", entry)
}

func (s *WebServe) Delete(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "DELETE", entry)
}

func (s *WebServe) Head(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "HEAD", entry)
}

func (s *WebServe) Option(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "OPTION", entry)
}

func (s *WebServe) Patch(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "PATCH", entry)
}

func (s *WebServe) All(pattern string, entry func(*Context)) bool {
	return s.router.AddEntry(pattern, "GET", entry) &&
		s.router.AddEntry(pattern, "POST", entry) &&
		s.router.AddEntry(pattern, "PUT", entry) &&
		s.router.AddEntry(pattern, "DELETE", entry) &&
		s.router.AddEntry(pattern, "HEAD", entry) &&
		s.router.AddEntry(pattern, "OPTION", entry) &&
		s.router.AddEntry(pattern, "PATCH", entry)
}

func (s *WebServe) Handle(entries interface{}) bool {
	return s.router.Handle(entries)
}

func (s *WebServe) Set404Page(f func(*Context)) {
	s.p404 = f
}