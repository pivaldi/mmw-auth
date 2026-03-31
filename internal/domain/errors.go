// Package user defines the auth bounded context's user aggregate and value objects.
package domain

import "errors"

var (
	// Validation errors
	ErrInvalidLogin    = errors.New("login is invalid")
	ErrInvalidPassword = errors.New("password cannot be empty")

	// Authentication errors
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid or expired token")

	// Resource errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
)
