package event_model

import (
	"errors"
	"time"

	aggregate_id_model "entrlcom.dev/event-sourcing/domain/model/aggregate/id"
	aggregate_type_model "entrlcom.dev/event-sourcing/domain/model/aggregate/type"
	aggregate_version_model "entrlcom.dev/event-sourcing/domain/model/aggregate/version"
	event_id_model "entrlcom.dev/event-sourcing/domain/model/event/id"
	event_timestamp_model "entrlcom.dev/event-sourcing/domain/model/event/timestamp"
	event_type_model "entrlcom.dev/event-sourcing/domain/model/event/type"
)

var ErrInvalidEvent = errors.New("invalid event")

type Event interface {
	GetAggregateID() aggregate_id_model.AggregateID
	GetAggregateType() aggregate_type_model.AggregateType
	GetAggregateVersion() aggregate_version_model.AggregateVersion
	GetData() []byte
	GetID() event_id_model.EventID
	GetTimestamp() event_timestamp_model.EventTimestamp
	GetType() event_type_model.EventType
	SetAggregateVersion(version aggregate_version_model.AggregateVersion)
}

type BaseEvent struct { //nolint:govet // OK.
	aggregateID      aggregate_id_model.AggregateID
	aggregateType    aggregate_type_model.AggregateType
	aggregateVersion aggregate_version_model.AggregateVersion
	data             []byte
	id               event_id_model.EventID
	timestamp        event_timestamp_model.EventTimestamp
	type_            event_type_model.EventType //nolint:revive // OK.
}

func (x BaseEvent) GetAggregateID() aggregate_id_model.AggregateID {
	return x.aggregateID
}

func (x BaseEvent) GetAggregateType() aggregate_type_model.AggregateType {
	return x.aggregateType
}

func (x BaseEvent) GetAggregateVersion() aggregate_version_model.AggregateVersion {
	return x.aggregateVersion
}

func (x BaseEvent) GetData() []byte {
	return x.data
}

func (x BaseEvent) GetID() event_id_model.EventID {
	return x.id
}

func (x BaseEvent) GetTimestamp() event_timestamp_model.EventTimestamp {
	return x.timestamp
}

func (x BaseEvent) GetType() event_type_model.EventType {
	return x.type_
}

func (x *BaseEvent) SetAggregateVersion(version aggregate_version_model.AggregateVersion) {
	x.aggregateVersion = version
}

func (x BaseEvent) Validate() error {
	if err := x.aggregateID.Validate(); err != nil {
		return errors.Join(err, ErrInvalidEvent)
	}

	if err := x.aggregateType.Validate(); err != nil {
		return errors.Join(err, ErrInvalidEvent)
	}

	if err := x.aggregateVersion.Validate(); err != nil {
		return errors.Join(err, ErrInvalidEvent)
	}

	if err := x.id.Validate(); err != nil {
		return errors.Join(err, ErrInvalidEvent)
	}

	if err := x.timestamp.Validate(); err != nil {
		return errors.Join(err, ErrInvalidEvent)
	}

	if err := x.type_.Validate(); err != nil {
		return errors.Join(err, ErrInvalidEvent)
	}

	return nil
}

func NewBaseEvent(type_ event_type_model.EventType, opts ...EventOption) (BaseEvent, error) {
	id, err := event_id_model.NewEventID()
	if err != nil {
		return BaseEvent{}, errors.Join(err, ErrInvalidEvent)
	}

	timestamp, err := event_timestamp_model.NewEventTimestamp(time.Now().UTC())
	if err != nil {
		return BaseEvent{}, errors.Join(err, ErrInvalidEvent)
	}

	x := BaseEvent{
		aggregateID:      "",
		aggregateType:    "",
		aggregateVersion: 0,
		data:             nil,
		id:               id,
		timestamp:        timestamp,
		type_:            type_,
	}

	for _, opt := range opts {
		if opt != nil {
			opt(&x)
		}
	}

	if err = x.Validate(); err != nil {
		return BaseEvent{}, err
	}

	return x, nil
}

type EventOption func(x Event)

func WithAggregate(
	aggregateID aggregate_id_model.AggregateID,
	aggregateType aggregate_type_model.AggregateType,
	aggregateVersion aggregate_version_model.AggregateVersion,
) EventOption {
	return func(x Event) {
		if event, ok := x.(*BaseEvent); ok {
			event.aggregateID = aggregateID
			event.aggregateType = aggregateType
			event.aggregateVersion = aggregateVersion
		}
	}
}

func WithData(data []byte) EventOption {
	return func(x Event) {
		if event, ok := x.(*BaseEvent); ok {
			event.data = data
		}
	}
}

func WithID(id event_id_model.EventID) EventOption {
	return func(x Event) {
		if event, ok := x.(*BaseEvent); ok {
			event.id = id
		}
	}
}

func WithTimestamp(timestamp event_timestamp_model.EventTimestamp) EventOption {
	return func(x Event) {
		if event, ok := x.(*BaseEvent); ok {
			event.timestamp = timestamp
		}
	}
}
