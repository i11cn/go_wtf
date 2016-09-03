package wtf

import ()

type (
	RespCodeItem struct {
		Describe string
		Content  func(*Context) []byte
	}
	RespCode map[int]RespCodeItem
)

func NewRespCode() RespCode {
	return make(RespCode)
}

func (r RespCode) SetRespCode(code int, desc string, content func(*Context) []byte) {
	r[code] = RespCodeItem{desc, content}
}

func (r RespCode) GetResp(code int) (bool, string, func(*Context) []byte) {
	if i, exist := r[code]; exist {
		return true, i.Describe, i.Content
	} else {
		return false, "", nil
	}
}
