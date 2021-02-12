package queue

import (
	"errors"
	"strings"

	"github.com/xh3b4sd/tracer"
)

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
