package consts

import "errors"

var (
	ErrRecordNotFound = errors.New("data not found")
	ErrConflict       = errors.New("single key breach")
	ErrInternalDB     = errors.New("internal database error.")
	ErrInternal       = errors.New("internal error in server.")
	ErrFieldNull      = errors.New("field is null")
)
