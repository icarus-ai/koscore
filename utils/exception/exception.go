package exception

import "fmt"

type IException interface{ Error() string }

type Exception struct{ message string }

func (m Exception) Error() string { return m.message }

func New(err string) IException { return Exception{message: err} }
func NewFormat(format string, args ...any) IException {
	return Exception{message: fmt.Sprintf(format, args)}
}

type (
	ArgumentOfRangeException IException
	NotSupportedException    IException
	OperationException       IException
	InvalidTargetException   IException
	UnmarshalException       IException
)

func NewArgumentOfRangeException(format string, args ...any) ArgumentOfRangeException {
	return ArgumentOfRangeException(NewFormat("argument of range exception: "+format, args))
}
func NewNotSupportedException(format string, args ...any) NotSupportedException {
	return NotSupportedException(NewFormat("not supported exception: "+format, args))
}
func NewOperationException(format string, args ...any) OperationException {
	return OperationException(NewFormat("operation exception: "+format, args))
}

func NewOperationExceptionCode[T int32 | uint32 | int64](code T, err string) OperationException {
	return OperationException(NewFormat("%s (%d)", err, code))
}

func NewInvalidTargetException(uin uint64, gin ...uint64) InvalidTargetException {
	if len(gin) == 0 {
		return InvalidTargetException(NewFormat("uid not found: uin %d", uin))
	}
	return InvalidTargetException(NewFormat("uid not found: gin %d uin %d", gin[0], uin))
}

type unmarshal_type string

const (
	UNMARSHAL_JSON  = "json"
	UNMARSHAL_PROTO = "proto"
)

func new_unmarshal_exception(typ unmarshal_type, tag string, err error) UnmarshalException {
	return UnmarshalException(NewFormat("unmarshal %s error: %s: %v", typ, tag, err))
}

func NewUnmarshalJsonException(err error, tag string) UnmarshalException {
	return new_unmarshal_exception(UNMARSHAL_JSON, tag, err)
}
func NewUnmarshalProtoException(err error, tag string) UnmarshalException {
	return new_unmarshal_exception(UNMARSHAL_PROTO, tag, err)
}
