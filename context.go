package wtf

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
)

type (
	// Context : 封装了上下文的结构
	Context struct {
		w         http.ResponseWriter
		r         *http.Request
		params    []UrlParams
		querys    map[string]string
		service   *WebService
		tpl       Template
		tpl_data  interface{}
		proc      func(*Context)
		mid_chain []MiddleWare
		body      []byte
		w_code    int
	}
)

func (c *Context) GetRequest() *http.Request {
	return c.r
}

func (c *Context) GetResponse() http.ResponseWriter {
	return c.w
}

func (c *Context) Process() {
	for _, chain := range c.mid_chain {
		if !chain(c) {
			return
		}
	}
	c.proc(c)
	c.ExecuteTemplate()
}

// LoadTemplateFiles 加载模板文件，可以同时加载多个相关的模板文件
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

// SetMime 设置输出的MIME类型
func (c *Context) SetMime(mime, charset string) {
	if len(charset) > 0 {
		c.w.Header().Set("Content-Type", fmt.Sprintf("%s;charset=%s", mime, charset))
	} else {
		c.w.Header().Set("Content-Type", mime)
	}
}

func (c *Context) WriteStatusCode(s int) {
	c.w.WriteHeader(s)
	c.w_code = s
	if page, exist := c.service.def_page[s]; exist {
		page(c)
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
		c.SetMime("application/json", "utf-8")
		_, err = c.Write(d)
	}
	return err
}

func (c *Context) WriteXml(obj interface{}) error {
	d, err := xml.Marshal(obj)
	if err == nil {
		c.SetMime("application/xml", "")
		_, err = c.Write(d)
	}
	return err
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

func (c *Context) GetJsonBody(o interface{}) error {
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

func (c *Context) GetParamByIndex(i int) string {
	if len(c.params) >= i {
		return c.params[i].Value
	}
	return ""
}
