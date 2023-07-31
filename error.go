package restype

type ErrorKind int

const (
	ErrorBodyFromRequest = iota
	ErrorExecuteRequest
	ErrorResponseFromBytes
)

func (e ErrorKind) Error() string {
	switch e {
	case ErrorBodyFromRequest:
		return "error obtaining body from request"
	case ErrorExecuteRequest:
		return "error executing request"
	case ErrorResponseFromBytes:
		return "error deserializing response"
	}
	return ""
}

type Error struct {
	kind     ErrorKind
	original error
}

func (e Error) Error() string {
	return e.kind.Error() + ": " + e.original.Error()
}

func (e Error) Is(target error) bool {
	_, ok := target.(Error)
	return ok
}

func (e Error) Unwrap() error {
	return e.original
}
