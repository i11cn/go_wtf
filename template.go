package wtf

import (
	"bytes"
	"html/template"
)

type (
	wtf_template struct {
		tpl *template.Template
	}
)

func NewTemplate() Template {
	return new_wtf_template()
}

func new_wtf_template() *wtf_template {
	ret := &wtf_template{}
	ret.tpl = template.New("wtf_root_tpl")
	return ret
}

func (wt *wtf_template) BindPipe(name string, f interface{}) {
	fm := map[string]interface{}{name: f}
	wt.tpl.Funcs(fm)
}

func (wt *wtf_template) LoadText(text string) {
	wt.tpl.Parse(text)
}

func (wt *wtf_template) LoadFiles(files ...string) {
	wt.tpl.ParseFiles(files...)
}

func (wt *wtf_template) Execute(name string, obj interface{}) ([]byte, error) {
	buf := &bytes.Buffer{}
	err := wt.tpl.ExecuteTemplate(buf, name, obj)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
