package daemon

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

var invalidFlagError = &tracer.Error{
	Kind: "invalidFlagError",
}

func IsInvalidFlag(err error) bool {
	return errors.Is(err, invalidFlagError)
}
