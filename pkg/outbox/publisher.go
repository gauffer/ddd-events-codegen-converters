package outbox

import (
	"fmt"
	"time"

	"github.com/gauffer/ddd-events-codegen-converters/pkg/domain"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/kafka"
	mdb "github.com/gauffer/ddd-events-codegen-converters/pkg/mongodb"

	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/protobuf/proto"
)

type outboxDocument struct {
	ID        string    `bson:"_id"`
	Topic     string    `bson:"topic"`
	Key       []byte    `bson:"key"`
	Message   []byte    `bson:"message"`
	CreatedAt time.Time `bson:"created_at"`
}

type Publisher struct {
	client     *mdb.Client
	database   string
	collection string
}

func NewPublisher(client *mdb.Client, database, collection string) *Publisher {
	return &Publisher{
		client:     client,
		database:   database,
		collection: collection,
	}
}

// Publish — сериализация доменных событий и запись в outbox
func (p *Publisher) Publish(sess mongo.SessionContext, events []domain.Event) error {
	if len(events) == 0 {
		return nil
	}

	docs := make([]any, 0, len(events))
	for _, e := range events {
		topic, err := kafka.GetTopic(e)
		if err != nil {
			return fmt.Errorf("get topic for %T: %w", e, err)
		}

		data, err := proto.Marshal(e)
		if err != nil {
			return fmt.Errorf("marshal %T: %w", e, err)
		}

		docs = append(docs, outboxDocument{
			ID:        e.EventID(),
			Topic:     topic,
			Key:       []byte(e.AggregateID()),
			Message:   data,
			CreatedAt: e.OccurredAtTime(),
		})
	}

	_, err := p.client.Database(p.database).Collection(p.collection).InsertMany(sess, docs)
	return err
}
