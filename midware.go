package wtf

import ()

type (
	MiddleWare func(c *Context) bool

	mid_chain_item struct {
		name string
		mid  MiddleWare
	}
)
