package store

import (
	"context"
	"fmt"

	"github.com/ovinc/zerotrust/internal/config"
	"github.com/ovinc/zerotrust/internal/otel"
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel/attribute"
)

var client *redis.Client

func init() {
	// init
	ctx := context.Background()

	// load config
	cfg := config.Get().Redis

	// create redis client with config
	client = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// enable tracing for redis client
	if err := redisotel.InstrumentTracing(
		client,
		redisotel.WithAttributes(
			attribute.String("db.instance", fmt.Sprintf("%s:%d/%d", cfg.Host, cfg.Port, cfg.DB)),
			attribute.String("db.ip", cfg.Host),
			attribute.String("db.system", "Redis"),
		),
	); err != nil {
		logrus.WithContext(ctx).Fatalf("failed to instrument redis tracing: %v", err)
	}

	// ping redis to verify connectivity
	if err := client.Ping(ctx).Err(); err != nil {
		logrus.WithContext(ctx).Fatalf("failed to ping redis: %v", err)
	}
}

func GetSession(ctx context.Context, sessionID string) (string, error) {
	// start new span
	ctx, span := otel.Tracer().Start(ctx, "store.redis.GetSession")
	defer span.End()

	// get session data from redis
	return client.Get(ctx, config.Get().Redis.FormatSessionKey(sessionID)).Result()
}

func Ping(ctx context.Context) error {
	return client.Ping(ctx).Err()
}

func Close() {
	if client != nil {
		_ = client.Close()
	}
}
