package wtf

import (
	"net/http"
)

type (
	wtf_resp_code_map struct {
		code_map map[int]func(Context)
	}
)

func NewResponseCode() ResponseCode {
	ret := &wtf_resp_code_map{}
	ret.code_map = make(map[int]func(Context))
	ret.code_map[http.StatusNotFound] = func(ctx Context) {
		ctx.WriteHeader(http.StatusNotFound)
		ctx.Write([]byte(ctx.Request().URL.Path))
		ctx.Write([]byte(": Not Found"))
	}
	return ret
}

func (rc *wtf_resp_code_map) Handle(code int, h func(Context)) {
	rc.code_map[code] = h
}

func (rc *wtf_resp_code_map) StatusCode(ctx Context, code int, body ...string) {
	if len(body) > 0 {
		ctx.WriteHeader(code)
		ctx.Write([]byte(body[0]))
	} else {
		if h, exist := rc.code_map[code]; exist {
			h(ctx)
		} else {
			ctx.WriteHeader(code)
		}
	}
}
