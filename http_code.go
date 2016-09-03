package wtf

import (
	"net/http"
)

type (
	RespCodeItem struct {
		Describe string
		Content  func(*http.Request)
	}
	RespCode map[int]RespCodeItem
)

func NewRespCode() RespCode {
	return make(RespCode)
}

func (r RespCode) SetRespCode(code int, desc string, content func(*http.Request)) {
	r[code] = RespCodeItem{desc, content}
}

func (r RespCode) GetResp(code int) (bool, string, func(*http.Request)) {
	if i, exist := r[code]; exist {
		return true, i.Describe, i.Content
	} else {
		return false, "", nil
	}
}
