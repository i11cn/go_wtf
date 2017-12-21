package wtf

import ()

type (
	wtf_template struct {
	}
)

func (wt *wtf_template) BindPipe(string) {
}

func (wt *wtf_template) Load(string, ...string) {
}

func (wt *wtf_template) LoadAll(...string) {
}

func (wt *wtf_template) Execute(string, interface{}) ([]byte, error) {
	return nil, nil
}
