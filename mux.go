package wtf

import ()

type (
	Mux interface {
		Handle(string, func(*Context)) error
		Default(func(*Context)) error
		HandleSubMux(string, Mux) error
		Match(*Context) func(*Context)
	}
)
