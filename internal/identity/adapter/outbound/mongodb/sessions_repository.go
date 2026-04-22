package mongodb

import (
	"context"
	"errors"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/model"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/repository"
	mdb "github.com/gauffer/ddd-events-codegen-converters/pkg/mongodb"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ repository.SessionsRepository = (*SessionsRepository)(nil)

type SessionsRepository struct {
	client *mdb.Client
}

type sessionDocument struct {
	ID               string     `bson:"_id"`
	Subject          string     `bson:"subject"`
	SubjectType      string     `bson:"subject_type"`
	ClientID         string     `bson:"client_id"`
	Scopes           []string   `bson:"scopes"`
	RefreshTokenHash string     `bson:"refresh_token_hash"`
	IP               string     `bson:"ip"`
	UserAgent        string     `bson:"user_agent"`
	CreatedAt        time.Time  `bson:"created_at"`
	LastUsedAt       time.Time  `bson:"last_used_at"`
	ExpiresAt        time.Time  `bson:"expires_at"`
	RevokedAt        *time.Time `bson:"revoked_at"`
}

func NewSessionsRepository(client *mdb.Client) repository.SessionsRepository {
	return &SessionsRepository{
		client: client,
	}
}

func (r *SessionsRepository) Create(ctx context.Context, session *model.Session) error {
	collection := r.collection()
	doc := r.sessionToDocument(session)

	_, err := collection.InsertOne(ctx, doc)
	return err
}

func (r *SessionsRepository) Save(ctx context.Context, session *model.Session) error {
	collection := r.collection()
	doc := r.sessionToDocument(session)

	filter := bson.M{"_id": string(session.ID)}
	opts := options.Replace().SetUpsert(true)

	_, err := collection.ReplaceOne(ctx, filter, doc, opts)
	return err
}

func (r *SessionsRepository) FindByID(ctx context.Context, id model.SessionID) (*model.Session, error) {
	collection := r.collection()
	filter := bson.M{"_id": string(id)}

	var doc sessionDocument
	if err := collection.FindOne(ctx, filter).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return r.documentToSession(&doc), nil
}

func (r *SessionsRepository) FindByRefreshTokenHash(ctx context.Context, hash string) (*model.Session, error) {
	collection := r.collection()
	filter := bson.M{"refresh_token_hash": hash}

	var doc sessionDocument
	if err := collection.FindOne(ctx, filter).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}

	return r.documentToSession(&doc), nil
}

func (r *SessionsRepository) collection() *mongo.Collection {
	return r.client.Database("app").Collection("sessions")
}

func (r *SessionsRepository) sessionToDocument(session *model.Session) *sessionDocument {
	scopes := make([]string, len(session.Scopes))
	for i, scope := range session.Scopes {
		scopes[i] = string(scope)
	}

	return &sessionDocument{
		ID:               string(session.ID),
		Subject:          session.Subject,
		SubjectType:      string(session.SubjectType),
		ClientID:         string(session.ClientID),
		Scopes:           scopes,
		RefreshTokenHash: session.RefreshTokenHash,
		IP:               session.IP,
		UserAgent:        session.UserAgent,
		CreatedAt:        session.CreatedAt,
		LastUsedAt:       session.LastUsedAt,
		ExpiresAt:        session.ExpiresAt,
		RevokedAt:        session.RevokedAt,
	}
}

func (r *SessionsRepository) documentToSession(doc *sessionDocument) *model.Session {
	scopes := make(model.OAuthScopes, len(doc.Scopes))
	for i, scope := range doc.Scopes {
		scopes[i] = model.OAuthScope(scope)
	}

	return &model.Session{
		ID:               model.SessionID(doc.ID),
		Subject:          doc.Subject,
		SubjectType:      model.SubjectType(doc.SubjectType),
		ClientID:         model.OAuthClientID(doc.ClientID),
		Scopes:           scopes,
		RefreshTokenHash: doc.RefreshTokenHash,
		IP:               doc.IP,
		UserAgent:        doc.UserAgent,
		CreatedAt:        doc.CreatedAt,
		LastUsedAt:       doc.LastUsedAt,
		ExpiresAt:        doc.ExpiresAt,
		RevokedAt:        doc.RevokedAt,
	}
}
