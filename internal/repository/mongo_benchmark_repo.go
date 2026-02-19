package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoBenchmarkRepository provides MongoDB SET/GET operations for benchmarking.
type MongoBenchmarkRepository struct {
	coll *mongo.Collection
}

// NewMongoBenchmarkRepository creates a new MongoBenchmarkRepository.
func NewMongoBenchmarkRepository(client *mongo.Client, db, collection string) *MongoBenchmarkRepository {
	return &MongoBenchmarkRepository{coll: client.Database(db).Collection(collection)}
}

// benchmarkDoc represents a document stored for benchmark (key=_id, value, expires_at).
type benchmarkDoc struct {
	ID        string    `bson:"_id"`
	Value     string    `bson:"value"`
	ExpiresAt time.Time `bson:"expires_at"`
}

// SetBenchmarkKey upserts a document with _id=key, value, and expires_at.
func (r *MongoBenchmarkRepository) SetBenchmarkKey(ctx context.Context, key string, value string, ttl time.Duration) error {
	expiresAt := time.Now().UTC().Add(ttl)
	doc := benchmarkDoc{ID: key, Value: value, ExpiresAt: expiresAt}

	opts := options.Replace().SetUpsert(true)
	_, err := r.coll.ReplaceOne(ctx, bson.M{"_id": key}, doc, opts)
	if err != nil {
		return fmt.Errorf("set benchmark key: %w", err)
	}
	return nil
}

// GetBenchmarkKey retrieves a document by _id and returns value and expires_at. Returns mongo.ErrNoDocuments if not found.
func (r *MongoBenchmarkRepository) GetBenchmarkKey(ctx context.Context, key string) (value string, expiresAt time.Time, err error) {
	var doc benchmarkDoc
	err = r.coll.FindOne(ctx, bson.M{"_id": key}).Decode(&doc)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("get benchmark key: %w", err)
	}
	return doc.Value, doc.ExpiresAt, nil
}

// DeleteBenchmarkKey deletes a document by _id. Used when simulating TTL for expired documents.
func (r *MongoBenchmarkRepository) DeleteBenchmarkKey(ctx context.Context, key string) error {
	_, err := r.coll.DeleteOne(ctx, bson.M{"_id": key})
	if err != nil {
		return fmt.Errorf("delete benchmark key: %w", err)
	}
	return nil
}
