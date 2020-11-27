package models

import "strings"

const (
	// ErrNotFound is returned when a resource cannot be found
	// in the database
	ErrNotFound modelError = "models: resource not found"
	// ErrInvalidID is returned when an invalid ID is provided
	// to the delete method
	ErrInvalidID modelError = "models: ID provided as invalid"
	// ErrInvalidPassword ...
	ErrInvalidPassword modelError = "models: invalid password provided"

	ErrPasswordTooShort modelError = "models: password must be at least 8 characters"

	ErrEmailRequired modelError = "models: email Address Is Required"

	ErrEmailInvalid modelError = "models: email Address is Not Valid"

	ErrEmailTaken modelError = "models: email address is already taken"

	ErrPasswordRequired modelError = "models: password is required"

	ErrRememberTooShort modelError = "models: remember token must be at least 32 bytes"
)

type modelError string

func (e modelError) Error() string {
	return string(e)
}

func (e modelError) Public() string {
	s := strings.Replace(string(e), "models: ", "", 1)
	return strings.Title(s)
}
