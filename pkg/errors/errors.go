package errors

import (
	"errors"
	"fmt"
)

// Common error codes
const (
	CodeNotFound         = "NOT_FOUND"
	CodeInvalidInput     = "INVALID_INPUT"
	CodeInternalError    = "INTERNAL_ERROR"
	CodeConflict         = "CONFLICT"
	CodeUnauthorized     = "UNAUTHORIZED"
	CodeValidationError  = "VALIDATION_ERROR"
	CodeTimeout          = "TIMEOUT"
	CodePermissionDenied = "PERMISSION_DENIED"
)

// Predefined error messages
const (
	MsgNotFound         = "the requested resource was not found"
	MsgInvalidInput     = "the input provided is invalid"
	MsgInternalError    = "an internal server error occurred"
	MsgConflict         = "a conflict occurred with the current state of the resource"
	MsgUnauthorized     = "unauthorized access"
	MsgValidationError  = "validation failed for the input data"
	MsgTimeout          = "the operation timed out"
	MsgPermissionDenied = "permission denied"
)

// Layer represents the layer in which the error occurred
type Layer string

const (
	LayerDelivery   Layer = "delivery"
	LayerService    Layer = "service"
	LayerRepository Layer = "repository"
)

// WrappedError represents an error with contextual information, including a stack of layers.
type WrappedError struct {
	LayerStack []Layer // The stack of layers through which the error passed
	Code       string  // An optional error code
	Message    string  // A human-readable message
	Original   error   // The original error
}

// Error implements the error interface, displaying the layer stack and the message.
func (e *WrappedError) Error() string {
	layerStackStr := fmt.Sprintf("[%s]", e.LayerStack[0])
	for i := 1; i < len(e.LayerStack); i++ {
		layerStackStr += fmt.Sprintf(" -> [%s]", e.LayerStack[i])
	}

	if e.Original != nil {
		return fmt.Sprintf("%s %s: %v", layerStackStr, e.Message, e.Original)
	}

	return fmt.Sprintf("%s %s", layerStackStr, e.Message)
}

// Unwrap allows WrappedError to be used with `errors.Unwrap` and `errors.Is`.
func (e *WrappedError) Unwrap() error {
	return e.Original
}

// New creates a new WrappedError without an original error, adding it to the start of the LayerStack.
func New(layer Layer, code, message string) error {
	return &WrappedError{
		LayerStack: []Layer{layer},
		Code:       code,
		Message:    message,
	}
}

// Wrap creates a new WrappedError by wrapping an existing error and adding the current layer to the stack.
func Wrap(err error, layer Layer, code, message string) error {
	if err == nil {
		return nil
	}

	var stack []Layer
	if wrappedErr, ok := err.(*WrappedError); ok {
		stack = append([]Layer{layer}, wrappedErr.LayerStack...)
	} else {
		stack = []Layer{layer}
	}

	return &WrappedError{
		LayerStack: stack,
		Code:       code,
		Message:    message,
		Original:   err,
	}
}

// Is checks if the error or any wrapped error matches the target.
func Is(err, target error) bool {
	return errors.Is(err, target)
}

// As checks if the error or any wrapped error can be cast to the target type.
func As(err error, target interface{}) bool {
	return errors.As(err, target)
}
