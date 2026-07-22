package domain

import "errors"

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrEmailAlreadyRegistered = errors.New("email is already registered")
	ErrInvalidEmailOrPassword = errors.New("invalid email or password")
	ErrTaskNotFound           = errors.New("task not found")
	ErrInvalidDateFormat      = errors.New("invalid due_date format, expected YYYY-MM-DD")
)
