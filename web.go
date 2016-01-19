package wtf

import (
	"encoding/json"
	"encoding/xml"
	"github.com/i11cn/go_logger"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

type (
	PathConfig struct {
		HtDoc        string
		TemplatePath string
		JsPath       string
		CSSPath      string
		ImagePath    string
	}

	UrlParams struct {
		Name  string
		Value string
	}

	Template interface {
		Load(...string) error
		Execute(interface{}) ([]byte, error)
	}

	Context struct {
		w          http.ResponseWriter
		r          *http.Request
		params     []UrlParams
		querys     map[string]string
		serve      *WebServe
		tpl        Template
		tpl_data   interface{}
		proc       func(*Context)
		mid_chain  []MiddleWare
		index      int
		chain_proc bool
		body       []byte
	}

	MiddleWare func(c *Context) bool

	Router interface {
		AddEntry(pattern string, method string, entry func(*Context)) bool
		Handle(entries interface{}) bool
		Match(url string, method string) (func(*Context), []UrlParams)
	}

	WebServe struct {
		path_config PathConfig
		router      Router
		fs_handler  http.Handler
		mid_chain   []mid_chain_item
		p404        func(*Context)
		p500        func(*Context)
	}
)

func init() {
	logger.GetLogger("web").AddAppender(logger.NewSplittedFileAppender("[%T] [%N-%L] %f@%F.%l: %M", "wtf.log", 24*time.Hour))
}

func (c *Context) GetRequest() *http.Request {
	return c.r
}

func (c *Context) GetResponse() http.ResponseWriter {
	return c.w
}

func (c *Context) Next() {
	if !c.chain_proc {
		return
	}
	if c.index >= len(c.mid_chain) {
		c.proc(c)
		c.chain_proc = false
		c.ExecuteTemplate()
	} else {
		c.index++
		c.chain_proc = c.mid_chain[c.index-1](c) && c.chain_proc
		c.Next()
	}
}

func (c *Context) LoadTemplateFiles(filenames ...string) error {
	return c.tpl.Load(filenames...)
}

func (c *Context) SetTemplateData(obj interface{}) {
	c.tpl_data = obj
}

func (c *Context) ExecuteTemplate() {
	data, err := c.tpl.Execute(c.tpl_data)
	if err == nil && len(data) > 0 {
		c.w.Write(data)
	}
	c.tpl_data = nil
}

func (c *Context) SetMime(mime string) {
	c.w.Header().Set("Content-Type", mime)
}

func (c *Context) GetBody() (string, error) {
	if len(c.body) > 0 {
		return string(c.body), nil
	}
	var err error
	c.body, err = ioutil.ReadAll(c.r.Body)
	if err != nil {
		return "", err
	} else {
		return string(c.body), nil
	}
}

func (c *Context) GetBodyAsJson(o interface{}) error {
	if len(c.body) < 1 {
		_, err := c.GetBody()
		if err != nil {
			return err
		}
	}
	return json.Unmarshal(c.body, o)
}

func (c *Context) GetQuery(name string) string {
	return c.querys[name]
}

func (c *Context) GetIntQuery(name string) (int64, bool) {
	v, exist := c.querys[name]
	if !exist {
		return 0, false
	}
	r, ok := strconv.ParseInt(v, 10, 64)
	return r, (ok == nil)
}

func (c *Context) GetParamByIndex(i int) string {
	if len(c.params) >= i {
		return c.params[i].Value
	}
	return ""
}

func (c *Context) GetIntParamByIndex(i int) (int64, bool) {
	if len(c.params) >= i {
		ret, err := strconv.ParseInt(c.params[i].Value, 10, 64)
		return ret, (err == nil)
	}
	return 0, false
}

func (c *Context) GetParam(name string) string {
	if len(name) > 0 {
		for _, s := range c.params {
			if s.Name == name {
				return s.Value
			}
		}
	}
	return ""
}

func (c *Context) GetIntParam(name string) (int64, bool) {
	if len(name) > 0 {
		for _, s := range c.params {
			if s.Name == name {
				ret, err := strconv.ParseInt(s.Value, 10, 64)
				return ret, (err == nil)
			}
		}
	}
	return 0, false
}

func (c *Context) WriteStatusCode(s int) {
	c.w.WriteHeader(s)
	if s == 404 {
		c.serve.p404(c)
	} else if s == 500 {
		c.serve.p500(c)
	}
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
		c.SetMime("application/json;charset=utf-8")
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

func NewWebServe(pc *PathConfig) *WebServe {
	ret := &WebServe{router: &default_router{}}
	if pc == nil {
		ret.path_config = PathConfig{HtDoc: "./htdoc", TemplatePath: "./template"}
	} else {
		ret.path_config = *pc
		if len(ret.path_config.HtDoc) < 1 {
			ret.path_config.HtDoc = "./htdoc"
		}
		if len(ret.path_config.TemplatePath) < 1 {
			ret.path_config.HtDoc = "./template"
		}
	}
	ret.p404 = func(c *Context) {
		c.WriteString("页面还在天上飞呢...是你在地上吹么？")
	}
	ret.p500 = func(c *Context) {
		c.WriteString("你干啥了？服务器都被你弄死了")
	}
	ret.fs_handler = http.FileServer(http.Dir(ret.path_config.HtDoc))
	ret.init()
	return ret
}

func (s *WebServe) SetRouter(r Router) {
	s.router = r
}

func (s *WebServe) GetRouter() Router {
	return s.router
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
	return s.router.AddEntry(pattern, "GET,POST,PUT,DELETE,HEAD,OPTION,PATCH", entry)
}

func (s *WebServe) Handle(entries interface{}) bool {
	return s.router.Handle(entries)
}

func (s *WebServe) Set404Page(f func(*Context)) {
	s.p404 = f
}

func (s *WebServe) AppendMiddleWare(f MiddleWare, name string) {
	for _, i := range s.mid_chain {
		if i.name == name {
			i.mid = f
			return
		}
	}
	s.mid_chain = append(s.mid_chain, mid_chain_item{name, f})
}

func (s *WebServe) DeleteMiddleWare(name string) {
	tmp := make([]mid_chain_item, 0, len(s.mid_chain))
	for _, i := range s.mid_chain {
		if i.name != name {
			tmp = append(tmp, i)
		}
	}
	s.mid_chain = tmp
}
