package event_type_model

import (
	"errors"
)

var ErrInvalidEventType = errors.New("invalid event type")

type EventType string

func (x EventType) String() string {
	return string(x)
}

func (x EventType) Validate() error {
	if len(x) == 0 {
		return ErrInvalidEventType
	}

	return nil
}

func NewEventType(s string) (EventType, error) {
	x := EventType(s)

	if err := x.Validate(); err != nil {
		return "", err
	}

	return x, nil
}
