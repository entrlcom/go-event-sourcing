package aggregate_model

import (
	"context"
	"errors"
)

var (
	ErrAggregateNotFound = errors.New("aggregate not found")
	ErrInternal          = errors.New("internal error")
)

type AggregateRepository interface {
	Load(ctx context.Context, aggregate Aggregate) error
	Save(ctx context.Context, aggregate Aggregate) error
}
