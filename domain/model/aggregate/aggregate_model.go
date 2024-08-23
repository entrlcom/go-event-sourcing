package aggregate_model

import (
	"errors"

	aggregate_id_model "entrlcom.dev/event-sourcing/domain/model/aggregate/id"
	aggregate_type_model "entrlcom.dev/event-sourcing/domain/model/aggregate/type"
	aggregate_version_model "entrlcom.dev/event-sourcing/domain/model/aggregate/version"
	event_model "entrlcom.dev/event-sourcing/domain/model/event"
)

type Aggregate interface {
	ApplyEvent(event event_model.Event) error
	ClearUncommittedEvents()
	GetAppliedEvents() []event_model.Event
	GetID() aggregate_id_model.AggregateID
	GetType() aggregate_type_model.AggregateType
	GetUncommittedEvents() []event_model.Event
	GetVersion() aggregate_version_model.AggregateVersion
	LoadEvents(events ...event_model.Event) error
	RaiseEvent(event event_model.Event) error
}

type when func(event event_model.Event) error

type BaseAggregate struct { //nolint:govet // OK.
	appliedEvents     []event_model.Event
	id                aggregate_id_model.AggregateID
	type_             aggregate_type_model.AggregateType
	uncommittedEvents []event_model.Event
	version           aggregate_version_model.AggregateVersion
	when              when
}

func (x *BaseAggregate) ApplyEvent(event event_model.Event) error {
	if !event.GetAggregateID().IsEqualTo(x.id) {
		return ErrInternal
	}

	if !event.GetAggregateType().IsEqualTo(x.type_) {
		return ErrInternal
	}

	if err := x.when(event); err != nil {
		return errors.Join(err, ErrInternal)
	}

	x.version++

	event.SetAggregateVersion(x.version)

	x.uncommittedEvents = append(x.uncommittedEvents, event)

	return nil
}

func (x *BaseAggregate) ClearUncommittedEvents() {
	x.uncommittedEvents = nil
}

func (x BaseAggregate) GetAppliedEvents() []event_model.Event {
	return x.appliedEvents
}

func (x BaseAggregate) GetID() aggregate_id_model.AggregateID {
	return x.id
}

func (x BaseAggregate) GetType() aggregate_type_model.AggregateType {
	return x.type_
}

func (x BaseAggregate) GetUncommittedEvents() []event_model.Event {
	return x.uncommittedEvents
}

func (x BaseAggregate) GetVersion() aggregate_version_model.AggregateVersion {
	return x.version
}

func (x *BaseAggregate) LoadEvents(events ...event_model.Event) error {
	for _, event := range events {
		if !event.GetAggregateID().IsEqualTo(x.id) {
			return ErrInternal
		}

		if err := x.when(event); err != nil {
			return errors.Join(err, ErrInternal)
		}

		x.appliedEvents = append(x.appliedEvents, event)
		x.version = event.GetAggregateVersion()
	}

	return nil
}

func (x *BaseAggregate) RaiseEvent(event event_model.Event) error {
	if !event.GetAggregateID().IsEqualTo(x.id) {
		return ErrInternal
	}

	if event.GetAggregateVersion() <= x.version {
		return ErrInternal
	}

	if err := x.when(event); err != nil {
		return errors.Join(err, ErrInternal)
	}

	x.appliedEvents = append(x.appliedEvents, event)
	x.version = event.GetAggregateVersion()

	return nil
}

func NewBaseAggregate(
	id aggregate_id_model.AggregateID,
	type_ aggregate_type_model.AggregateType,
	opts ...BaseAggregateOption,
) (BaseAggregate, error) {
	x := BaseAggregate{
		appliedEvents:     nil,
		id:                id,
		type_:             type_,
		uncommittedEvents: nil,
		version:           aggregate_version_model.DefaultAggregateVersion,
		when:              func(event event_model.Event) error { return nil },
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&x)
		}
	}

	return x, nil
}

type BaseAggregateOption func(x *BaseAggregate)

func WithWhen(fn when) BaseAggregateOption {
	return func(x *BaseAggregate) {
		x.when = fn
	}
}

var _ Aggregate = (*BaseAggregate)(nil)
