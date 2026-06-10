package model

type AppError struct {
	StatusCode       int
	ErrorCategory    string
	ErrorDescription string
}
