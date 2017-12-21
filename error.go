package wtf

type (
	wtf_error struct {
		code int
		msg  string
	}
)

func NewError(code int, msg string) Error {
	return &wtf_error{code, msg}
}

func trans_error(code int, err error) Error {
	if err != nil {
		return NewError(code, err.Error())
	} else {
		return nil
	}
}

func (e wtf_error) Code() int {
	return e.code
}

func (e wtf_error) Message() string {
	return e.msg
}
