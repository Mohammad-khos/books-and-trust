package domain

import "errors"

var (
	ErrRegexNotMatched  = errors.New("password does not meet security requirements: it must contain at least one uppercase letter, one number, one special character, and be at least 8 characters long")
	ErrValidationFailed = errors.New("failed to validate request payload")
	ErrInvalidToken = errors.New("invalid token")
	ErrResourceAleadyExists = errors.New("resource already exists")
	ErrInvalidCredential = errors.New("invalid credential")
	ErrResourceNotFound = errors.New("resource not found")
	ErrNoFieldsToUpdate = errors.New("no fields to update")
)
