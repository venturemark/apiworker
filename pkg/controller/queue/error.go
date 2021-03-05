package queue

import (
	"errors"
	"strings"

	"github.com/xh3b4sd/tracer"
)

var incompleteExecutionError = &tracer.Error{
	Kind: "incompleteExecutionError",
	Desc: "This error indicates that the execution of a task could not be completed successfully. Tasks should be rescheduled after incomplete execution in order to finish them accordingly.",
}

func IsIncompleteExecution(err error) bool {
	return errors.Is(err, incompleteExecutionError)
}

var invalidConfigError = &tracer.Error{
	Kind: "invalidConfigError",
}

func IsInvalidConfig(err error) bool {
	return errors.Is(err, invalidConfigError)
}

var dialError = &tracer.Error{
	Kind: "dialError",
}

func IsDialError(err error) bool {
	if err == nil {
		return false
	}

	{
		s := tracer.Cause(err).Error()

		if strings.Contains(s, "EOF") {
			return true
		}
		if strings.Contains(s, "dial tcp") {
			return true
		}
		if strings.Contains(s, "read tcp") {
			return true
		}
		if strings.Contains(s, "connection refused") {
			return true
		}
		if strings.Contains(s, "connection reset") {
			return true
		}
	}

	return errors.Is(err, dialError)
}
