package wtf

import ()

type (
	Mux interface {
		Handle(string, func(*Context))
		Match(string) (func(*Context), bool)
	}
)
