package profile

import (
	profileMongo "github.com/gauffer/ddd-events-codegen-converters/internal/profile/adapter/outbound/mongodb"
	"github.com/gauffer/ddd-events-codegen-converters/internal/profile/domain/repository"
	mdb "github.com/gauffer/ddd-events-codegen-converters/pkg/mongodb"
)

type Adapters struct {
	profiles repository.ProfilesRepository
	mongo    *mdb.Client
}

type AdaptersOptions func(*Adapters)

func WithMongoClient(client *mdb.Client) AdaptersOptions {
	return func(a *Adapters) { a.mongo = client }
}

func NewAdapters(options ...AdaptersOptions) *Adapters {
	a := &Adapters{}
	for _, opt := range options {
		opt(a)
	}

	if a.mongo == nil {
		panic("mongo client is required")
	}

	a.profiles = profileMongo.NewProfilesRepository(a.mongo)

	return a
}

func (a *Adapters) GetProfilesRepository() repository.ProfilesRepository {
	return a.profiles
}
