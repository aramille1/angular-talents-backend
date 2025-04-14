package internal

type CustomError struct {
	status int
	code string
	message string
	detail string
}

func NewError(status int, code, message, detail string) (*CustomError) {
	newErr := &CustomError{
		status, 
		code,
		message,
		detail,
	}

	return newErr
}

func (e *CustomError) Error() string {
	return e.message
}

func (e *CustomError) ErrorData() map[string]interface{} {
	return map[string]interface{}{
		"status": e.status,
		"code": e.code,
		"message": e.message,
		"detail": e.detail,
	}
}