package identity

// Расширение proto-сгенерированного UserCreatedEvent методами domain.Event (cosmos-sdk подход)

import "time"

func (e *UserCreatedEvent) EventID() string {
	return e.GetEventId()
}

func (e *UserCreatedEvent) EventType() string {
	return string(e.ProtoReflect().Descriptor().FullName())
}

func (e *UserCreatedEvent) AggregateID() string {
	return e.GetUserId()
}

func (e *UserCreatedEvent) OccurredAtTime() time.Time {
	return e.GetOccurredAt().AsTime()
}
