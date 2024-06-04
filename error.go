package main

import "fmt"

type Error struct {
	Token   Token
	Message string
	Line    int
}

func (e *Error) Error() string {
	return fmt.Sprintf("[Line: %d] Error: %s\n", e.Line, e.Message)
}

// Error handle struct to manager errors
type ErrorHandler struct {
	Errors []*Error
}

func NewErrorHandler() *ErrorHandler {
	return &ErrorHandler{
		Errors: make([]*Error, 0),
	}
}

func (eh *ErrorHandler) AddError(err *Error) {
	eh.Errors = append(eh.Errors, err)
}

func (eh *ErrorHandler) HasErrors() bool {
	return len(eh.Errors) > 0
}

func (eh *ErrorHandler) PrintErrors() {
	for _, err := range eh.Errors {
		fmt.Println(err)
	}
}
