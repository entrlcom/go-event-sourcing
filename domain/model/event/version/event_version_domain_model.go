package event_version_model

import (
	"errors"
)

var ErrInvalidEventVersion = errors.New("invalid event version")

type EventVersion int64

func (x EventVersion) Validate() error {
	if x < 0 {
		return ErrInvalidEventVersion
	}

	return nil
}
