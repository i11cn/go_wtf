package wtf

import (
	"bytes"
	"html/template"
	"net/http"
)

type (
	wtf_template struct {
		tpl *template.Template
	}
)

func NewTemplate() Template {
	ret := &wtf_template{}
	ret.tpl = template.New("wtf_root_tpl")
	return ret
}

func (wt *wtf_template) BindPipe(name string, f interface{}) Template {
	fm := map[string]interface{}{name: f}
	wt.tpl.Funcs(fm)
	return wt
}

func (wt *wtf_template) LoadText(text string) Template {
	wt.tpl.Parse(text)
	return wt
}

func (wt *wtf_template) LoadFiles(files ...string) Template {
	wt.tpl.ParseFiles(files...)
	return wt
}

func (wt *wtf_template) Execute(name string, obj interface{}) ([]byte, Error) {
	buf := &bytes.Buffer{}
	err := wt.tpl.ExecuteTemplate(buf, name, obj)
	if err != nil {
		return nil, NewError(http.StatusInternalServerError, "模板执行错误", err)
	}
	return buf.Bytes(), nil
}
