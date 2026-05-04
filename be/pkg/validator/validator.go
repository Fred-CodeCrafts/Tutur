// Package validator provides lightweight input validation helpers used across
// all HTTP handlers. It collects field-level errors and returns them in a
// structured format compatible with the standard error response envelope.
package validator

import (
	"fmt"
	"strings"
)

// ValidationError holds a map of field → error message pairs.
type ValidationError struct {
	Fields map[string]string
}

func (e *ValidationError) Error() string {
	parts := make([]string, 0, len(e.Fields))
	for field, msg := range e.Fields {
		parts = append(parts, fmt.Sprintf("%s: %s", field, msg))
	}
	return strings.Join(parts, "; ")
}

// Validator accumulates validation errors for a single request.
type Validator struct {
	errors map[string]string
}

// New returns a fresh Validator.
func New() *Validator {
	return &Validator{errors: make(map[string]string)}
}

// Check adds an error for field if condition is false.
func (v *Validator) Check(condition bool, field, message string) {
	if !condition {
		v.AddError(field, message)
	}
}

// AddError records an error for the given field (first error wins).
func (v *Validator) AddError(field, message string) {
	if _, exists := v.errors[field]; !exists {
		v.errors[field] = message
	}
}

// Valid returns true when no errors have been recorded.
func (v *Validator) Valid() bool {
	return len(v.errors) == 0
}

// Errors returns the accumulated field errors.
func (v *Validator) Errors() map[string]string {
	return v.errors
}

// Err returns a *ValidationError if there are errors, or nil if valid.
func (v *Validator) Err() error {
	if v.Valid() {
		return nil
	}
	return &ValidationError{Fields: v.errors}
}
