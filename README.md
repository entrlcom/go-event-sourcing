# Event Sourcing

## Table of Content

- [Examples](#examples)
- [License](#license)

## Examples

```go
package main

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/EventStore/EventStore-Client-Go/v4/esdb"
	"github.com/google/uuid"

	aggregate_model "entrlcom.dev/event-sourcing/domain/model/aggregate"
	aggregate_id_model "entrlcom.dev/event-sourcing/domain/model/aggregate/id"
	event_model "entrlcom.dev/event-sourcing/domain/model/event"
	event_id_model "entrlcom.dev/event-sourcing/domain/model/event/id"
	event_timestamp_model "entrlcom.dev/event-sourcing/domain/model/event/timestamp"
	eventstore_repository "entrlcom.dev/event-sourcing/infra/repository/aggregate/eventstore"
)

func main() {
	ctx := context.Background()

	// EventStore configuration.
	eventStoreDBConfiguration, err := esdb.ParseConnectionString("esdb://localhost:2113/?tls=false")
	if err != nil {
		return
	}
	
	// EventStore client.
	eventStoreDBClient, err := esdb.NewClient(eventStoreDBConfiguration)
	if err != nil {
		return
	}

	// Repository.
	repository := eventstore_repository.NewEventStoreAggregateRepository(*eventStoreDBClient)

	// New account (aggregate).
	account, err := NewAccount(NewAccountID())
	if err != nil {
		return
	}

	// Load account (aggregate).
	if err = repository.Load(ctx, account._aggregate); err != nil {
		if !errors.Is(err, aggregate_model.ErrAggregateNotFound) {
			return
		}
	}

	if err = account.ChangeName("A"); err != nil {
		return
	}

	if err = account.ChangeName("B"); err != nil {
		return
	}

	// Save account (aggregate).
	if err = repository.Save(ctx, account._aggregate); err != nil {
		return
	}

	if err = account.ChangeName("C"); err != nil {
		return
	}

	if err = account.ChangeName("D"); err != nil {
		return
	}

	// Save account (aggregate).
	if err = repository.Save(ctx, account._aggregate); err != nil {
		return
	}
}

var ErrInvalidAccount = errors.New("invalid account")

const AggregateType = "Account"

type Account struct {
	_aggregate *aggregate_model.BaseAggregate

	id   AccountID
	name string
}

func (x *Account) ChangeName(name string) error {
	event := NewAccountNameChangedEvent(x.id.String(), name)

	data, _ := json.Marshal(event)

	eventID, err := event_id_model.NewEventID()
	if err != nil {
		return err
	}

	timestamp, err := event_timestamp_model.NewEventTimestamp(time.Now().UTC())
	if err != nil {
		return err
	}

	e, err := event_model.NewBaseEvent(
		eventID,
		AccountNameChangedEventType,
		event_model.WithAggregate(
			x._aggregate.GetID(),
			x._aggregate.GetType(),
			x._aggregate.GetVersion(),
		),
		event_model.WithData(data),
		event_model.WithTimestamp(timestamp),
	)
	if err != nil {
		return err
	}

	if err = x._aggregate.ApplyEvent(&e); err != nil {
		return err
	}

	return nil
}

func NewAccount(id AccountID) (*Account, error) {
	x := Account{
		_aggregate: nil,
		id:         id,
		name:       "",
	}

	aggregateID, err := aggregate_id_model.NewAggregateIDFromString(id.String())
	if err != nil {
		return nil, errors.Join(err, ErrInvalidAccount)
	}

	aggregate, err := aggregate_model.NewBaseAggregate(
		aggregateID,
		AggregateType,
		aggregate_model.WithWhen(func(event event_model.Event) error {
			switch event.GetType() {
			case AccountNameChangedEventType:
				var data AccountNameChangedEvent

				if err = json.Unmarshal(event.GetData(), &data); err != nil {
					return err
				}

				x.name = data.Name

				return nil
			default:
				return nil
			}
		}),
	)
	if err != nil {
		return nil, errors.Join(err, ErrInvalidAccount)
	}

	x._aggregate = &aggregate

	return &x, nil
}

var ErrInvalidAccountID = errors.New("invalid account id")

type AccountID string

func (x AccountID) IsEqualTo(accountID AccountID) bool {
	return x == accountID
}

func (x AccountID) String() string {
	return string(x)
}

func (x AccountID) Validate() error {
	v, err := uuid.Parse(string(x))
	if err != nil || v.Variant() != uuid.RFC4122 || v.Version() != 7 {
		return ErrInvalidAccountID
	}

	return nil
}

func NewAccountID() AccountID {
	return AccountID(uuid.Must(uuid.NewV7()).String())
}

const AccountNameChangedEventType = "AccountNameChanged"

type AccountNameChangedEvent struct {
	AccountID string `json:"account_id,omitempty"`
	Name      string `json:"name,omitempty"`
}

func NewAccountNameChangedEvent(accountID, name string) AccountNameChangedEvent {
	return AccountNameChangedEvent{
		AccountID: accountID,
		Name:      name,
	}
}

```

## License

[MIT](https://choosealicense.com/licenses/mit/)
