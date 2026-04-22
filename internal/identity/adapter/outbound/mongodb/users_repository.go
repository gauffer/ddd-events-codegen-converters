package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/repository"
	mdb "github.com/gauffer/ddd-events-codegen-converters/pkg/mongodb"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/outbox"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ repository.UserRepository = (*UsersRepository)(nil)

type UsersRepository struct {
	client *mdb.Client
	outbox *outbox.Publisher
}

type userDocument struct {
	ID          string     `bson:"_id"`
	PublicID    *int64     `bson:"public_id"`
	Phone       string     `bson:"phone"`
	Status      string     `bson:"status"`
	CreatedAt   time.Time  `bson:"created_at"`
	UpdatedAt   time.Time  `bson:"updated_at"`
	LastLoginAt *time.Time `bson:"last_login_at"`
}

func NewUsersRepository(client *mdb.Client, outbox *outbox.Publisher) repository.UserRepository {
	return &UsersRepository{
		client: client,
		outbox: outbox,
	}
}

func (r *UsersRepository) Create(ctx context.Context, user *model.User) error {
	return r.client.WithTransaction(ctx, func(sess mongo.SessionContext) (any, error) {
		doc := r.userToDocument(user)

		result, err := r.collection().InsertOne(sess, doc)
		if err != nil {
			return nil, fmt.Errorf("insert user: %w", err)
		}

		if user.HasEvents() {
			if err := r.outbox.Publish(sess, user.PopEvents()); err != nil {
				return nil, fmt.Errorf("publish to outbox: %w", err)
			}
		}

		return result.InsertedID, nil
	})
}

func (r *UsersRepository) Save(ctx context.Context, user *model.User) error {
	filter := bson.M{"_id": string(user.ID)}
	opts := options.Replace().SetUpsert(true)
	_, err := r.collection().ReplaceOne(ctx, filter, r.userToDocument(user), opts)
	return err
}

func (r *UsersRepository) FindByID(ctx context.Context, id model.UserID) (*model.User, error) {
	var doc userDocument
	if err := r.collection().FindOne(ctx, bson.M{"_id": string(id)}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return r.documentToUser(&doc), nil
}

func (r *UsersRepository) FindByPhone(ctx context.Context, phone model.Phone) (*model.User, error) {
	var doc userDocument
	if err := r.collection().FindOne(ctx, bson.M{"phone": string(phone)}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return r.documentToUser(&doc), nil
}

func (r *UsersRepository) collection() *mongo.Collection {
	return r.client.Database("app").Collection("users")
}

func (r *UsersRepository) userToDocument(user *model.User) *userDocument {
	return &userDocument{
		ID:          string(user.ID),
		PublicID:    user.PublicID,
		Phone:       string(user.Phone),
		Status:      string(user.Status),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		LastLoginAt: user.LastLoginAt,
	}
}

func (r *UsersRepository) documentToUser(doc *userDocument) *model.User {
	return &model.User{
		ID:          model.UserID(doc.ID),
		PublicID:    doc.PublicID,
		Phone:       model.Phone(doc.Phone),
		Status:      model.UserStatus(doc.Status),
		CreatedAt:   doc.CreatedAt,
		UpdatedAt:   doc.UpdatedAt,
		LastLoginAt: doc.LastLoginAt,
	}
}
