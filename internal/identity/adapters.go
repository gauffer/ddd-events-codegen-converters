package identity

import (
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/adapter/outbound/inmemory"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/adapter/outbound/mongodb"
	redisAdapter "github.com/gauffer/ddd-events-codegen-converters/internal/identity/adapter/outbound/redis"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/application/port"
	"github.com/gauffer/ddd-events-codegen-converters/internal/identity/domain/repository"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"
	mdb "github.com/gauffer/ddd-events-codegen-converters/pkg/mongodb"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/outbox"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/service"

	"github.com/redis/go-redis/v9"
)

type Adapters struct {
	otpSender port.OTPSender
	jwtSigner port.JWTSigner

	authorizationCodes repository.AuthorizationCodesRepository
	sessions           repository.SessionsRepository
	oauthClients       repository.OAuthClientsRepository
	users              repository.UserRepository
	otpSessions        repository.OTPSessionsRepository

	mongoClient *mdb.Client
	redisClient *redis.Client
	outbox      *outbox.Publisher
}

type AdaptersOptions func(*Adapters)

func WithOTPSender(s port.OTPSender) AdaptersOptions { return func(a *Adapters) { a.otpSender = s } }
func WithJWTSigner(s port.JWTSigner) AdaptersOptions  { return func(a *Adapters) { a.jwtSigner = s } }
func WithMongoClient(c *mdb.Client) AdaptersOptions      { return func(a *Adapters) { a.mongoClient = c } }
func WithRedisClient(c *redis.Client) AdaptersOptions    { return func(a *Adapters) { a.redisClient = c } }
func WithOutboxPublisher(o *outbox.Publisher) AdaptersOptions { return func(a *Adapters) { a.outbox = o } }

func NewAdapters(_ service.Environment, _ log.Logger, options ...AdaptersOptions) *Adapters {
	a := &Adapters{}
	for _, opt := range options {
		opt(a)
	}

	a.authorizationCodes = redisAdapter.NewAuthCodesRepository(a.redisClient)
	a.otpSessions = redisAdapter.NewOTPSessionsRepository(a.redisClient)
	a.sessions = mongodb.NewSessionsRepository(a.mongoClient)
	a.users = mongodb.NewUsersRepository(a.mongoClient, a.outbox)
	a.oauthClients = inmemory.NewOAuthClientsRepository()

	return a
}
