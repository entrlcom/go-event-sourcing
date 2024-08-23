package aggregate_version_model

import (
	"errors"
)

const DefaultAggregateVersion = -1

var ErrInvalidAggregateVersion = errors.New("invalid aggregate version")

type AggregateVersion int64

func (x AggregateVersion) Validate() error {
	if x < DefaultAggregateVersion {
		return ErrInvalidAggregateVersion
	}

	return nil
}

func NewAggregateVersionFromUint64(aggregateVersion uint64) (AggregateVersion, error) {
	x := AggregateVersion(aggregateVersion)

	if err := x.Validate(); err != nil {
		return 0, err
	}

	return x, nil
}
