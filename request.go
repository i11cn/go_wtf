package wtf

import (
	"io/ioutil"
	"net/http"
)

type (
	RestParams interface {
		Get(string) string
		GetByIndex(int) string
	}

	Request struct {
		*http.Request
		RestParams
		Uri  string
		Body []byte
	}
)

func NewRequest(req *http.Request) (*Request, error) {
	ret := &Request{}
	ret.Request = req
	ret.RestParams = &empty_rest_params{}
	ret.Uri = req.URL.Path
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		return nil, err
	}
	ret.Body = body
	return ret, nil
}

type (
	empty_rest_params struct {
	}
)

func (*empty_rest_params) Get(string) string {
	return ""
}

func (*empty_rest_params) GetByIndex(int) string {
	return ""
}
