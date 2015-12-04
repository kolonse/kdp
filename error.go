// error
package kdp

import (
	"errors"
	"strconv"
)

/**
*	错误码定义
 */
const (
	KDP_PROTO_SUCCESS      = iota
	KDP_PROTO_ERROR_PARAM  = 10001
	KDP_PROTO_ERROR_LENGTH // 长度错误
	KDP_PROTO_ERROR_FORMAT // 格式错误
	KDP_PROTO_ERROR_NET    = 20001
	KDP_PROTO_ERROR_LOGIC  = 30001
	KDP_PROTO_ERROR_NOT_KDP_PROTO
)

type Error struct {
	err  error
	code int
}

func (e *Error) Error() string {
	return "Code:" + strconv.Itoa(e.code) + " Error:" + e.err.Error()
}

func (e *Error) GetCode() int {
	return e.code
}

func NewError(code int, message string) Error {
	return Error{
		code: code,
		err:  errors.New(message),
	}
}
