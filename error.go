package wtf

import (
	"fmt"
)

type (
	wtf_error struct {
		err  error
		code int
		msg  string
	}
)

func NewError(code int, msg string, err ...error) Error {
	ret := &wtf_error{}
	ret.code, ret.msg = code, msg
	if len(err) > 0 {
		ret.err = err[0]
	}
	return ret
}

func trans_error(code int, err error) Error {
	if err != nil {
		return NewError(code, err.Error())
	} else {
		return nil
	}
}

func (e wtf_error) Error() string {
	if e.err != nil {
		return e.err.Error()
	} else {
		return fmt.Sprintf("Error Code %d : %s", e.code, e.msg)
	}
}

func (e wtf_error) Code() int {
	return e.code
}

func (e wtf_error) Message() string {
	return e.msg
}
