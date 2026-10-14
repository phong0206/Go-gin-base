package errors

type Kind string

const (
	KindValidation   Kind = "validation"
	KindUnauthorized Kind = "unauthorized"
	KindForbidden    Kind = "forbidden"
	KindNotFound     Kind = "not_found"
	KindConflict     Kind = "conflict"
	KindMethod       Kind = "method_not_allowed"
	KindInternal     Kind = "internal"
)

type AppError struct {
	Kind    Kind
	Code    string
	Message string
	Err     error
}

func (e *AppError) Error() string {
	if e.Err == nil {
		return e.Message
	}
	return e.Message + ": " + e.Err.Error()
}

func (e *AppError) Unwrap() error {
	return e.Err
}

func New(kind Kind, code, message string) *AppError {
	return &AppError{Kind: kind, Code: code, Message: message}
}

func Wrap(kind Kind, code, message string, err error) *AppError {
	return &AppError{Kind: kind, Code: code, Message: message, Err: err}
}

var ErrUserNotFound = New(KindNotFound, "USER_NOT_FOUND", "user not found")
