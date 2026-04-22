package app

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	identity "github.com/gauffer/ddd-events-codegen-converters/internal/identity/adapter/inbound/http"
	"github.com/gauffer/ddd-events-codegen-converters/internal/pkg/jws"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"
	"github.com/gauffer/ddd-events-codegen-converters/pkg/middleware"
	mdb "github.com/gauffer/ddd-events-codegen-converters/pkg/mongodb"
	svc "github.com/gauffer/ddd-events-codegen-converters/pkg/service"

	"github.com/justinas/alice"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

func newServer(
	addr string,
	shutdownTimeout time.Duration,
	logger log.Logger,
	identityHandlers *identity.IdentityHandlers,
	jwtValidator jws.Validator,
	mongoClient *mdb.Client,
	redisClient *redis.Client,
) *svc.HTTPServer {
	mux := http.NewServeMux()

	identity.HandlerFromMux(identityHandlers, mux)
	mux.HandleFunc("GET /v1/demo/db-check", dbCheckHandler(mongoClient))
	mux.HandleFunc("POST /v1/demo/reset", resetHandler(mongoClient, redisClient))

	chain := alice.New(
		middleware.RequestID(),
		middleware.Logger(logger),
		middleware.Recover(logger, middleware.LogPanicRequest),
		middleware.ContentType("application/json"),
		jwtWithSubject(jwtValidator, logger),
	)

	return svc.NewHTTPServer(addr, shutdownTimeout, chain.Then(mux), logger)
}

func dbCheckHandler(client *mdb.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		db := client.Database("app")

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"users":    fetchAll(ctx, db, "users"),
			"profiles": fetchAll(ctx, db, "profiles"),
			"outbox":   fetchAll(ctx, db, "identity_outbox"),
		})
	}
}

func resetHandler(client *mdb.Client, redisClient *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		db := client.Database("app")
		for _, coll := range []string{"users", "profiles", "identity_outbox"} {
			_, _ = db.Collection(coll).DeleteMany(ctx, bson.M{})
		}
		_ = redisClient.FlushDB(ctx).Err()

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok"})
	}
}

func fetchAll(ctx context.Context, db *mongo.Database, coll string) []bson.M {
	cursor, err := db.Collection(coll).Find(ctx, bson.M{})
	if err != nil {
		return nil
	}
	var results []bson.M
	_ = cursor.All(ctx, &results)
	return results
}
