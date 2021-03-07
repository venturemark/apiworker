package roledelete

import (
	"errors"

	"github.com/xh3b4sd/tracer"
)

var invalidConfigError = &tracer.Error{
	Kind: "invalidConfigError",
}

func IsInvalidConfig(err error) bool {
	return errors.Is(err, invalidConfigError)
}

var invalidTaskError = &tracer.Error{
	Kind: "invalidTaskError",
}

func IsInvalidTask(err error) bool {
	return errors.Is(err, invalidTaskError)
}

var timeoutError = &tracer.Error{
	Kind: "timeoutError",
}

func IsTimeout(err error) bool {
	return errors.Is(err, timeoutError)
}
