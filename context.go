package wtf

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
)

type (
	UrlParams struct {
		Name  string
		Value string
	}

	Context struct {
		w          http.ResponseWriter
		r          *http.Request
		params     []UrlParams
		querys     map[string]string
		serve      *WebService
		tpl        Template
		tpl_data   interface{}
		proc       func(*Context)
		mid_chain  []MiddleWare
		chain_proc bool
		body       []byte
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

func (c *Context) WriteStatusCode(s int) {
	c.w.WriteHeader(s)
	if page, exist := c.serve.def_page[s]; exist {
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
