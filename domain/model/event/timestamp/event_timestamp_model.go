package event_timestamp_model

import (
	"errors"
	"time"
)

var ErrInvalidEventTimestamp = errors.New("invalid event timestamp")

type EventTimestamp time.Time

func (x EventTimestamp) Validate() error {
	if t := time.Time(x); t.IsZero() || t.After(time.Now().UTC()) {
		return ErrInvalidEventTimestamp
	}

	return nil
}

func NewEventTimestamp(t time.Time) (EventTimestamp, error) {
	x := EventTimestamp(t)

	if err := x.Validate(); err != nil {
		return EventTimestamp{}, err
	}

	return x, nil
}
