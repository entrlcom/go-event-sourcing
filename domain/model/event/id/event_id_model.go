package event_id_model

import (
	"errors"

	"github.com/google/uuid"
)

var ErrInvalidEventID = errors.New("invalid event id")

type EventID uuid.UUID

func (x EventID) IsEqualTo(id EventID) bool {
	return x == id
}

func (x EventID) UUID() uuid.UUID {
	return uuid.UUID(x)
}

func (x EventID) Validate() error {
	if eventID := uuid.UUID(x); eventID == uuid.Nil || eventID.Variant() != uuid.RFC4122 {
		return ErrInvalidEventID
	}

	return nil
}

func NewEventID() (EventID, error) {
	eventID, err := uuid.NewV7()
	if err != nil {
		return EventID{}, errors.Join(err, ErrInvalidEventID)
	}

	return EventID(eventID), nil
}

func NewEventIDFromUUID(eventID uuid.UUID) (EventID, error) {
	x := EventID(eventID)

	if err := x.Validate(); err != nil {
		return EventID{}, err
	}

	return x, nil
}
