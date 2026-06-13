package domain

import "errors"

var (
	ErrNotFound                 = errors.New("resource not found")
	ErrNoFieldsToUpdate         = errors.New("no fields to update")
	ErrInvalidFieldsToUpdate    = errors.New("invalid fields to update")
	ErrUserIsOnBannedUsers      = errors.New("user is currently banned")
	ErrInvalidID                = errors.New("invalid or empty uuid provided")
	ErrInvalidDeliveryCode      = errors.New("invalid or empty delivery code")
	ErrPermissionDenied         = errors.New("permission denied")
	ErrSelfOperationNotAllowed  = errors.New("operation on self is not allowed")
	ErrValidation               = errors.New("validation failed")
	ErrLoanAlreadyTaken         = errors.New("cannot update a loan that is already active or claimed by a user")
	ErrLoanAlreadyBeenDelivered = errors.New("the loan has already been delivered")
	ErrLoanReturned             = errors.New("this loan has been returned and is no longer available")
)
