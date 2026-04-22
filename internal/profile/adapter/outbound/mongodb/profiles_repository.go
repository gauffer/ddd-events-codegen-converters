package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/gauffer/ddd-events-codegen-converters/internal/profile/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/internal/profile/domain/repository"
	mdb "github.com/gauffer/ddd-events-codegen-converters/pkg/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ repository.ProfilesRepository = (*ProfilesRepository)(nil)

type ProfilesRepository struct {
	client *mdb.Client
}

type profileDocument struct {
	UserID    string    `bson:"_id"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

func NewProfilesRepository(client *mdb.Client) repository.ProfilesRepository {
	return &ProfilesRepository{client: client}
}

func (r *ProfilesRepository) FindByUserID(ctx context.Context, userID string) (*model.Profile, error) {
	var doc profileDocument
	if err := r.collection().FindOne(ctx, bson.M{"_id": userID}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return &model.Profile{
		ID:        model.ProfileID(doc.UserID),
		CreatedAt: doc.CreatedAt,
		UpdatedAt: doc.UpdatedAt,
	}, nil
}

func (r *ProfilesRepository) Save(ctx context.Context, profile *model.Profile) error {
	doc := profileDocument{
		UserID:    string(profile.ID),
		CreatedAt: profile.CreatedAt,
		UpdatedAt: profile.UpdatedAt,
	}

	filter := bson.M{"_id": string(profile.ID)}
	update := bson.M{"$set": doc}
	opts := options.Update().SetUpsert(true)

	_, err := r.collection().UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *ProfilesRepository) collection() *mongo.Collection {
	return r.client.Database("app").Collection("profiles")
}
