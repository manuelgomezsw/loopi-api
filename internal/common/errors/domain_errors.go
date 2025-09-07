package errors

type DomainError interface {
	error
	Status() int
	Message() string
}

type domainError struct {
	status  int
	message string
}

func NewDomainError(status int, message string) DomainError {
	return &domainError{
		status:  status,
		message: message,
	}
}

func (e *domainError) Error() string   { return e.message }
func (e *domainError) Status() int     { return e.status }
func (e *domainError) Message() string { return e.message }
