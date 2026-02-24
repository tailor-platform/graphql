package gqlerrors

import (
	"errors"

	"github.com/tailor-platform/graphql/language/location"
)

type ExtendedError interface {
	error
	Extensions() map[string]interface{}
}

type FormattedError struct {
	Message       string                    `json:"message"`
	Locations     []location.SourceLocation `json:"locations"`
	Path          []interface{}             `json:"path,omitempty"`
	Extensions    map[string]interface{}    `json:"extensions,omitempty"`
	originalError error
}

func (g FormattedError) OriginalError() error {
	return g.originalError
}

func (g FormattedError) Error() string {
	return g.Message
}

func NewFormattedError(message string) FormattedError {
	err := errors.New(message)
	return FormatError(err)
}

func FormatError(err error) FormattedError {
	switch err := err.(type) {
	case FormattedError:
		return err
	case *Error:
		ret := FormattedError{
			Message:       err.Error(),
			Locations:     err.Locations,
			Path:          err.Path,
			originalError: err,
		}
		if err := err.OriginalError; err != nil {
			if extended, ok := err.(ExtendedError); ok {
				ret.Extensions = extended.Extensions()
			}
		}
		return ret
	case Error:
		return FormatError(&err)
	default:
		return FormattedError{
			Message:       err.Error(),
			Locations:     []location.SourceLocation{},
			originalError: err,
		}
	}
}

func FormatErrors(errs ...error) []FormattedError {
	formattedErrors := []FormattedError{}
	for _, err := range errs {
		formattedErrors = append(formattedErrors, FormatError(err))
	}
	return formattedErrors
}

func FormatErrorsFromError(err error) []FormattedError {
	gqlErr, isGqlErr := err.(*Error)

	if isGqlErr && gqlErr.OriginalError != nil {
		if unwrapper, ok := gqlErr.OriginalError.(interface{ Unwrap() []error }); ok {
			var result []FormattedError
			for _, e := range unwrapper.Unwrap() {
				newErr := &Error{
					Message:       e.Error(),
					Locations:     gqlErr.Locations,
					Path:          gqlErr.Path,
					OriginalError: e,
				}
				result = append(result, FormatErrorsFromError(newErr)...)
			}
			return result
		}
	}

	if unwrapper, ok := err.(interface{ Unwrap() []error }); ok {
		var result []FormattedError
		for _, e := range unwrapper.Unwrap() {
			result = append(result, FormatErrorsFromError(e)...)
		}
		return result
	}

	return []FormattedError{FormatError(err)}
}
