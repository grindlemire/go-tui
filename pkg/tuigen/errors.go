package tuigen

import (
	"fmt"
	"strings"
)

// Error represents a compilation error with source location and optional hint.
type Error struct {
	Pos     Position
	Message string
	Hint    string // optional suggestion for fixing the error
}

// Error implements the error interface.
func (e *Error) Error() string {
	var sb strings.Builder
	sb.WriteString(e.Pos.String())
	sb.WriteString(": error: ")
	sb.WriteString(e.Message)
	if e.Hint != "" {
		sb.WriteString(" (")
		sb.WriteString(e.Hint)
		sb.WriteString(")")
	}
	return sb.String()
}

// NewError creates a new Error with the given position and message.
func NewError(pos Position, message string) *Error {
	return &Error{Pos: pos, Message: message}
}

// NewErrorf creates a new Error with a formatted message.
func NewErrorf(pos Position, format string, args ...any) *Error {
	return &Error{Pos: pos, Message: fmt.Sprintf(format, args...)}
}

// NewErrorWithHint creates a new Error with a hint for fixing the error.
func NewErrorWithHint(pos Position, message, hint string) *Error {
	return &Error{Pos: pos, Message: message, Hint: hint}
}

// ErrorList collects multiple errors during compilation.
type ErrorList struct {
	errors []*Error
}

// NewErrorList creates an empty error list.
func NewErrorList() *ErrorList {
	return &ErrorList{}
}

// Add appends an error to the list.
func (el *ErrorList) Add(err *Error) {
	el.errors = append(el.errors, err)
}

// AddError creates and adds an error with the given position and message.
func (el *ErrorList) AddError(pos Position, message string) {
	el.errors = append(el.errors, NewError(pos, message))
}

// AddErrorf creates and adds an error with a formatted message.
func (el *ErrorList) AddErrorf(pos Position, format string, args ...any) {
	el.errors = append(el.errors, NewErrorf(pos, format, args...))
}

// Len returns the number of errors.
func (el *ErrorList) Len() int {
	return len(el.errors)
}

// HasErrors returns true if there are any errors.
func (el *ErrorList) HasErrors() bool {
	return len(el.errors) > 0
}

// Errors returns a copy of the error slice.
func (el *ErrorList) Errors() []*Error {
	result := make([]*Error, len(el.errors))
	copy(result, el.errors)
	return result
}

// Error implements the error interface, returning all errors joined by newlines.
func (el *ErrorList) Error() string {
	if len(el.errors) == 0 {
		return ""
	}
	if len(el.errors) == 1 {
		return el.errors[0].Error()
	}

	var sb strings.Builder
	for i, err := range el.errors {
		if i > 0 {
			sb.WriteByte('\n')
		}
		sb.WriteString(err.Error())
	}
	return sb.String()
}

// Err returns nil if there are no errors, otherwise returns the ErrorList as an error.
func (el *ErrorList) Err() error {
	if len(el.errors) == 0 {
		return nil
	}
	return el
}
