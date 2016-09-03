package wtf

import ()

type (
	// Context : 封装了上下文的结构
	Context struct {
		*Request
		Response
		Template
	}
)
