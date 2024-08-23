package aggregate_type_model

import (
	"errors"
)

var ErrInvalidAggregateType = errors.New("invalid aggregate type")

type AggregateType string

func (x AggregateType) IsEqualTo(aggregateType AggregateType) bool {
	return x == aggregateType
}

func (x AggregateType) String() string {
	return string(x)
}

func (x AggregateType) Validate() error {
	if len(x) == 0 {
		return ErrInvalidAggregateType
	}

	return nil
}

func NewAggregateType(s string) (AggregateType, error) {
	x := AggregateType(s)

	if err := x.Validate(); err != nil {
		return "", err
	}

	return x, nil
}
