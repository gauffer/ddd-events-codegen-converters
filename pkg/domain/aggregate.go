package domain

import (
	"time"

	"google.golang.org/protobuf/proto"
)

// Event — доменное событие на базе proto. Метаданные добавляются в _ext.go
type Event interface {
	proto.Message
	EventID() string
	EventType() string
	AggregateID() string
	OccurredAtTime() time.Time
}

type AggregateRoot struct {
	events []Event
}

func (a *AggregateRoot) AddEvent(event Event) { a.events = append(a.events, event) }
func (a *AggregateRoot) HasEvents() bool       { return len(a.events) > 0 }
func (a *AggregateRoot) PopEvents() []Event {
	out := a.events
	a.events = nil
	return out
}
