package httperrors

import "fmt"

type StatusError struct {
	Code int
	Err  error
}

func (se StatusError) Error() string {
	return se.Err.Error()
}

func New(code int, err error) StatusError {
	return StatusError{Code: code, Err: err}
}

func Newf(code int, format string, a ...any) StatusError {
	return StatusError{Code: code, Err: fmt.Errorf(format, a...)}
}
