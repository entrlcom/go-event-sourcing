package aggregate_id_model

import (
	"errors"
)

var ErrInvalidAggregateID = errors.New("invalid aggregate id")

type AggregateID string

func (x AggregateID) IsEqualTo(id AggregateID) bool {
	return x == id
}

func (x AggregateID) String() string {
	return string(x)
}

func (x AggregateID) Validate() error {
	if len(x) == 0 {
		return ErrInvalidAggregateID
	}

	return nil
}

func NewAggregateIDFromString(s string) (AggregateID, error) {
	x := AggregateID(s)

	if err := x.Validate(); err != nil {
		return "", err
	}

	return x, nil
}
