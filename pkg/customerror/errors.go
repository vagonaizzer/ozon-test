package customerror

import "fmt"


type NotFoundError struct {
	Entity string
	ID     string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s with id %q not found", e.Entity, e.ID)
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field %q: %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}


type ForbiddenError struct {
	Reason string
}

func (e *ForbiddenError) Error() string {
	return fmt.Sprintf("forbidden: %s", e.Reason)
}


type ConflictError struct {
	Message string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("conflict: %s", e.Message)
}


func NotFound(entity, id string) error {
	return &NotFoundError{Entity: entity, ID: id}
}


func Forbidden(reason string) error {
	return &ForbiddenError{Reason: reason}
}


func Validation(field, msg string) error {
	return &ValidationError{Field: field, Message: msg}
}
