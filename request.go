package wtf

import (
	"net/http"
)

type (
	Request interface {
		Method() string
		AuthInfo() (user, pass string, ok bool)
		Cookie(name string) (*http.Cookie, error)
		Cookies() []*http.Cookie
		Referer() string
		Host() string
		Uri() string
		Body() []byte
	}
)
