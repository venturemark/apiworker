package remindercreate

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

var timeoutError = &tracer.Error{
	Kind: "timeoutError",
}

func IsTimeout(err error) bool {
	return errors.Is(err, timeoutError)
}

var mailDeliveryError = &tracer.Error{
	Kind: "mailDeliveryError",
}

func IsMailDelivery(err error) bool {
	return errors.Is(err, mailDeliveryError)
}
