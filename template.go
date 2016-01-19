package wtf

import (
	. "bytes"
	"fmt"
	"html/template"
)

type (
	Template interface {
		Load(...string) error
		Execute(interface{}) ([]byte, error)
	}

	default_template struct {
		path string
		tpl  *template.Template
	}
)

func (t *default_template) Load(filenames ...string) (err error) {
	fn := make([]string, len(filenames))
	for i, f := range filenames {
		fn[i] = fmt.Sprintf("%s/%s", t.path, f)
	}
	if t.tpl == nil {
		t.tpl, err = template.ParseFiles(fn...)
	} else {
		_, err = t.tpl.ParseFiles(fn...)
	}
	return
}

func (t *default_template) Execute(obj interface{}) (ret []byte, err error) {
	if t.tpl == nil {
		return
	}
	var buf Buffer
	err = t.tpl.Execute(&buf, obj)
	if err == nil {
		ret = buf.Bytes()
	}
	t.tpl = nil
	return
}
