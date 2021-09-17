package joberrors

import (
//	"errors"
)

//go:generate stringer -type=Error_Type
type Error_Type int;
const (
	Error_None Error_Type = iota
	Error_Panic
	Error_ConcurrentJob
	Error_DeferedJob
	Error_Middleware
)

type FailedJobError struct {
	Message string
	ErrorType Error_Type
}

func (e FailedJobError) Error() string {
	return e.Message;
}

func (e FailedJobError) Is(target error) bool {
	_, ok := target.(FailedJobError)
	if ok == false {
		return false
	}
	return true;
}

//ErrorScheduleNotFound Error When we can't find a Schedule
type ErrorScheduleNotFound struct {
	Message string
}

func (e ErrorScheduleNotFound) Error() string {
	return e.Message
}

//ErrorScheduleExists Error When a schedule already exists
type ErrorScheduleExists struct {
	Message string
}

func (e ErrorScheduleExists) Error() string {
	return e.Message
}


