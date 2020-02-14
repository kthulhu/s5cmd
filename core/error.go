package core

import (
	"errors"
	"strings"
)

// AcceptableError interface defines an error which is OK-to-have, for things like "cp -n" etc. It should not be treated as an error (regarding the exit code etc)
type AcceptableError interface {
	error
	Acceptable() bool
}

// AcceptableErrorType embeds the stdlib error interface so that we can have more receivers on it
type AcceptableErrorType struct {
	error
}

// NewAcceptableError creates a new AcceptableError
func NewAcceptableError(s string) AcceptableErrorType {
	return AcceptableErrorType{errors.New(s)}
}

// Acceptable is always true for errors of AcceptableError type
func (e AcceptableErrorType) Acceptable() bool {
	return true
}

// CleanupError converts multiline error messages generated by aws-sdk-go into a single line
func CleanupError(err error) (s string) {
	s = strings.Replace(err.Error(), "\n", " ", -1)
	s = strings.Replace(s, "\t", " ", -1)
	s = strings.Replace(s, "  ", " ", -1)
	s = strings.TrimSpace(s)
	return
}

// IsAcceptableError determines if the error is an AcceptableError, and if so, returns the error as such
func IsAcceptableError(err error) AcceptableError {
	e, ok := err.(AcceptableError)
	if !ok {
		return nil
	}
	return e
}
