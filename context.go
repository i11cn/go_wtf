package wtf

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"net/http"
	"strconv"
)

type (
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
)

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
