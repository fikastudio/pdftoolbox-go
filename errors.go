package pdftoolbox

type ReasonCode uint16
type ReturnCode uint8
type ErrorCode uint8

type Error struct {
	Code    ErrorCode
	Message string
}

func newError(code ErrorCode) Error {
	e := Error{
		Code:    code,
		Message: errors[code],
	}

	return e
}

const CodeNotSerialized ErrorCode = 100

var errors = map[ErrorCode]string{
	CodeNotSerialized: "Not serialized (no valid serialization found or keycode expired)",
}
