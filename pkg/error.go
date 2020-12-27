package pkg

import "fmt"

type InternalErrorType string

const (
	NotImplemented InternalErrorType = "NOT_IMPLEMENTED"
)

type InternalError struct {
	Message   string
	ErrorType InternalErrorType
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("%s (%s)", e.Message, e.ErrorType)
}

func CreateNotImplementedError(message string) error {
	return CreateInternalError(message, NotImplemented)
}

func CreateInternalError(
	message string,
	errorType InternalErrorType,
) error {
	return &InternalError{
		Message:   message,
		ErrorType: errorType,
	}
}
