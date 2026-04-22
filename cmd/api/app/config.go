package app

import (
	"os"
	"time"
)

type Config struct {
	RestAddress        string
	DiagnosticsAddress string
	ShutdownTimeout    time.Duration
	MongoURI           string
	RedisURI           string
	JWTPrivateKeyB64   string
	AppEnv             string

	KafkaBrokers      string
	KafkaSASLEnabled  bool
	KafkaSASLUser     string
	KafkaSASLPassword string
}

func NewConfig() *Config {
	return &Config{
		RestAddress:        envOrDefault("REST_ADDRESS", ":8080"),
		DiagnosticsAddress: envOrDefault("DIAGNOSTICS_ADDRESS", ":7073"),
		ShutdownTimeout:    parseDuration(envOrDefault("SHUTDOWN_TIMEOUT", "10s")),
		MongoURI:           envOrDefault("MONGO_URI", "mongodb://root:password@localhost:27017/?replicaSet=rs0&authSource=admin"),
		RedisURI:           envOrDefault("REDIS_URI", "redis://:yourpassword@localhost:6379"),
		JWTPrivateKeyB64:   os.Getenv("JWT_PRIVATE_KEY_B64"),
		AppEnv:             envOrDefault("APP_ENV", "local"),

		KafkaBrokers:      envOrDefault("KAFKA_BROKERS", "localhost:19092"),
		KafkaSASLEnabled:  envOrDefault("KAFKA_SASL_ENABLED", "false") == "true",
		KafkaSASLUser:     os.Getenv("KAFKA_SASL_USER"),
		KafkaSASLPassword: os.Getenv("KAFKA_SASL_PASSWORD"),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 10 * time.Second
	}
	return d
}
