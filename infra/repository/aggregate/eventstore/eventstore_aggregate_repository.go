package eventstore_repository

import (
	"context"
	"errors"
	"io"
	"math"
	"strings"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"

	aggregate_model "entrlcom.dev/event-sourcing/domain/model/aggregate"
	aggregate_id_model "entrlcom.dev/event-sourcing/domain/model/aggregate/id"
	aggregate_type_model "entrlcom.dev/event-sourcing/domain/model/aggregate/type"
	aggregate_version_model "entrlcom.dev/event-sourcing/domain/model/aggregate/version"
	event_model "entrlcom.dev/event-sourcing/domain/model/event"
	event_id_model "entrlcom.dev/event-sourcing/domain/model/event/id"
	event_timestamp_model "entrlcom.dev/event-sourcing/domain/model/event/timestamp"
	event_type_model "entrlcom.dev/event-sourcing/domain/model/event/type"
)

type EventStoreAggregateRepository struct {
	eventStoreDBClient esdb.Client
}

//nolint:cyclop,gocognit // OK.
func (x EventStoreAggregateRepository) Load(ctx context.Context, aggregate aggregate_model.Aggregate) error {
	if aggregate == nil {
		return aggregate_model.ErrInternal
	}

	streamID, err := NewStreamID(aggregate.GetID(), aggregate.GetType())
	if err != nil {
		return errors.Join(err, aggregate_model.ErrInternal)
	}

	readStreamOptions := esdb.ReadStreamOptions{} //nolint:exhaustruct // OK.

	readStream, err := x.eventStoreDBClient.ReadStream(ctx, streamID.String(), readStreamOptions, math.MaxInt64)
	if err != nil {
		return errors.Join(err, aggregate_model.ErrInternal)
	}

	defer readStream.Close()

	for {
		resolvedEvent, err := readStream.Recv()
		if err != nil { //nolint:nestif // OK.
			if errors.Is(err, io.EOF) {
				break
			}

			if err, ok := esdb.FromError(err); !ok {
				if err.Code() == esdb.ErrorCodeResourceNotFound {
					return errors.Join(err, aggregate_model.ErrAggregateNotFound)
				}
			}

			return errors.Join(err, aggregate_model.ErrInternal)
		}

		event, err := NewEventFromRecordedEvent(resolvedEvent.OriginalEvent())
		if err != nil {
			return errors.Join(err, aggregate_model.ErrInternal)
		}

		if err = aggregate.RaiseEvent(event); err != nil {
			return errors.Join(err, aggregate_model.ErrInternal)
		}
	}

	return nil
}

func (x EventStoreAggregateRepository) Save(ctx context.Context, aggregate aggregate_model.Aggregate) error {
	if aggregate == nil {
		return aggregate_model.ErrInternal
	}

	uncommittedEvents := aggregate.GetUncommittedEvents()
	if len(uncommittedEvents) == 0 {
		return nil
	}

	if err := x.save(ctx, aggregate); err != nil {
		return errors.Join(err, aggregate_model.ErrInternal)
	}

	aggregate.ClearUncommittedEvents()

	// Load events in case the aggregate needs to be further used after this save.
	if err := aggregate.LoadEvents(uncommittedEvents...); err != nil {
		return errors.Join(err, aggregate_model.ErrInternal)
	}

	return nil
}

func (x EventStoreAggregateRepository) new(ctx context.Context, aggregate aggregate_model.Aggregate) error {
	eventsData, err := eventsToEventStoreEventsData(aggregate.GetUncommittedEvents())
	if err != nil {
		return err
	}

	streamID, err := NewStreamID(aggregate.GetID(), aggregate.GetType())
	if err != nil {
		return err
	}

	appendToStreamOptions := esdb.AppendToStreamOptions{
		ExpectedRevision: esdb.NoStream{},
		Authenticated:    nil,
		Deadline:         nil,
		RequiresLeader:   false,
	}

	if _, err = x.eventStoreDBClient.AppendToStream(
		ctx,
		streamID.String(),
		appendToStreamOptions,
		eventsData...,
	); err != nil {
		return err
	}

	return nil
}

func (x EventStoreAggregateRepository) save(ctx context.Context, aggregate aggregate_model.Aggregate) error {
	streamID, err := NewStreamID(aggregate.GetID(), aggregate.GetType())
	if err != nil {
		return err
	}

	readStreamOptions := esdb.ReadStreamOptions{
		Direction:      esdb.Backwards,
		From:           esdb.End{},
		ResolveLinkTos: false,
		Authenticated:  nil,
		Deadline:       nil,
		RequiresLeader: false,
	}

	readStream, err := x.eventStoreDBClient.ReadStream(ctx, streamID.String(), readStreamOptions, 1)
	if err != nil {
		return err
	}

	defer readStream.Close()

	resolvedEvent, err := readStream.Recv()
	if err != nil { //nolint:nestif // OK.
		if err, ok := esdb.FromError(err); !ok {
			if err.Code() == esdb.ErrorCodeResourceNotFound {
				return x.new(ctx, aggregate)
			}
		}

		return err
	}

	eventsData, err := eventsToEventStoreEventsData(aggregate.GetUncommittedEvents())
	if err != nil {
		return err
	}

	appendToStreamOptions := esdb.AppendToStreamOptions{
		ExpectedRevision: esdb.Revision(resolvedEvent.OriginalEvent().EventNumber),
		Authenticated:    nil,
		Deadline:         nil,
		RequiresLeader:   false,
	}

	if _, err = x.eventStoreDBClient.AppendToStream(
		ctx,
		streamID.String(),
		appendToStreamOptions,
		eventsData...,
	); err != nil {
		return err
	}

	return nil
}

func NewEventStoreAggregateRepository(eventStoreDBClient esdb.Client) EventStoreAggregateRepository {
	return EventStoreAggregateRepository{eventStoreDBClient: eventStoreDBClient}
}

func NewEventFromRecordedEvent(recordedEvent *esdb.RecordedEvent) (event_model.Event, error) {
	if recordedEvent == nil {
		return nil, event_model.ErrInvalidEvent
	}

	streamID, err := NewStreamIDFromString(recordedEvent.StreamID)
	if err != nil {
		return nil, errors.Join(err, event_model.ErrInvalidEvent)
	}

	aggregateVersion, err := aggregate_version_model.NewAggregateVersionFromUint64(recordedEvent.EventNumber)
	if err != nil {
		return nil, errors.Join(err, event_model.ErrInvalidEvent)
	}

	eventID, err := event_id_model.NewEventIDFromUUID(recordedEvent.EventID)
	if err != nil {
		return nil, errors.Join(err, event_model.ErrInvalidEvent)
	}

	eventTimestamp, err := event_timestamp_model.NewEventTimestamp(recordedEvent.CreatedDate)
	if err != nil {
		return nil, errors.Join(err, event_model.ErrInvalidEvent)
	}

	eventType, err := event_type_model.NewEventType(recordedEvent.EventType)
	if err != nil {
		return nil, errors.Join(err, event_model.ErrInvalidEvent)
	}

	event, err := event_model.NewBaseEvent(
		eventID,
		eventType,
		event_model.WithAggregate(
			streamID.GetAggregateID(),
			streamID.GetAggregateType(),
			aggregateVersion,
		),
		event_model.WithData(recordedEvent.Data),
		event_model.WithTimestamp(eventTimestamp),
	)
	if err != nil {
		return nil, errors.Join(err, event_model.ErrInvalidEvent)
	}

	return &event, nil
}

var ErrInvalidStreamID = errors.New("invalid event stream id")

const separator = ":"

type StreamID struct {
	aggregateID   aggregate_id_model.AggregateID
	aggregateType aggregate_type_model.AggregateType
}

func (x StreamID) GetAggregateID() aggregate_id_model.AggregateID {
	return x.aggregateID
}

func (x StreamID) GetAggregateType() aggregate_type_model.AggregateType {
	return x.aggregateType
}

func (x StreamID) String() string {
	return x.aggregateType.String() + separator + x.aggregateID.String()
}

func (x StreamID) Validate() error {
	if err := x.aggregateID.Validate(); err != nil {
		return errors.Join(err, ErrInvalidStreamID)
	}

	if err := x.aggregateType.Validate(); err != nil {
		return errors.Join(err, ErrInvalidStreamID)
	}

	return nil
}

func NewStreamID(
	aggregateID aggregate_id_model.AggregateID,
	aggregateType aggregate_type_model.AggregateType,
) (StreamID, error) {
	x := StreamID{
		aggregateID:   aggregateID,
		aggregateType: aggregateType,
	}

	if err := x.Validate(); err != nil {
		return StreamID{}, err
	}

	return x, nil
}

func NewStreamIDFromString(s string) (StreamID, error) {
	v := strings.SplitN(s, separator, 2) //nolint:gomnd,mnd // OK.

	aggregateID, err := aggregate_id_model.NewAggregateIDFromString(v[1])
	if err != nil {
		return StreamID{}, err
	}

	aggregateType, err := aggregate_type_model.NewAggregateType(v[0])
	if err != nil {
		return StreamID{}, err
	}

	return NewStreamID(aggregateID, aggregateType)
}

func eventsToEventStoreEventsData(events []event_model.Event) ([]esdb.EventData, error) {
	switch l := len(events); l {
	case 0:
		return nil, nil
	case 1:
		eventData, err := eventToEventStoreEventData(events[0])
		if err != nil {
			return nil, err
		}

		return []esdb.EventData{eventData}, nil
	default:
		eventsData := make([]esdb.EventData, 0, len(events))

		for i := range events {
			eventData, err := eventToEventStoreEventData(events[i])
			if err != nil {
				return nil, err
			}

			eventsData = append(eventsData, eventData)
		}

		return eventsData, nil
	}
}

func eventToEventStoreEventData(event event_model.Event) (esdb.EventData, error) {
	if event == nil {
		return esdb.EventData{}, event_model.ErrInvalidEvent
	}

	return esdb.EventData{
		EventID:     event.GetID().UUID(),
		EventType:   event.GetType().String(),
		ContentType: esdb.ContentTypeJson,
		Data:        event.GetData(),
		Metadata:    nil,
	}, nil
}

var _ aggregate_model.AggregateRepository = (*EventStoreAggregateRepository)(nil)
