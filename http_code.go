package wtf

type (
	RespCodeItem struct {
		Describe string
		Content  func(*Context)
	}
	RespCode map[int]RespCodeItem
)

func NewRespCode() RespCode {
	return make(RespCode)
}

func (r RespCode) SetRespCode(code int, desc string, content func(*Context)) {
	m := map[int]RespCodeItem(r)
	m[code] = RespCodeItem{desc, content}
}

func (r RespCode) GetResp(code int) (bool, string, func(*Context)) {
	m := map[int]RespCodeItem(r)
	if i, exist := m[code]; exist {
		return true, i.Describe, i.Content
	} else {
		return false, "", nil
	}
}
