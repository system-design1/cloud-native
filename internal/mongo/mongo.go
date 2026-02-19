package mongo

import (
	"context"
	"fmt"
	"time"

	cfg "go-backend-service/internal/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// NewClient creates a MongoDB client with pool and timeout settings from config and verifies connectivity with a ping.
func NewClient(c cfg.MongoConfig) (*mongo.Client, error) {
	opts := options.Client().
		ApplyURI(c.URI).
		SetMaxPoolSize(c.MaxPoolSize).
		SetMinPoolSize(c.MinPoolSize).
		SetConnectTimeout(c.ConnectTimeout).
		SetServerSelectionTimeout(c.ServerSelectionTimeout).
		SetSocketTimeout(c.SocketTimeout).
		SetHeartbeatInterval(c.HeartbeatInterval)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		return nil, fmt.Errorf("mongo connect failed: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping failed: %w", err)
	}

	return client, nil
}
