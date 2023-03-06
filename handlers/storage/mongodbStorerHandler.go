package storage

import (
	"context"
	"errors"

	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

//var log = logger.GetOrCreate("storage")

//var errKeyNotFound = errors.New("key not found")

// ErrNilRedisClientWrapper signals that a nil mongodb client has been provided
var ErrNilMongoDBClient = errors.New("nil mongodb client provided")

type mongodbStorerHandler struct {
	client mongodb.MongoDBClientWrapper
	coll   *mongo.Collection
	ctx    context.Context
}

// NewMongoDBStorerHandler will create a new storer handler instance
func NewMongoDBStorerHandler(client mongodb.MongoDBClientWrapper) (*mongodbStorerHandler, error) {
	if client == nil {
		return nil, ErrNilMongoDBClient
	}

	ctx := context.Background()

	// begin create and insert
	coll := client.GetCollection("users")

	return &mongodbStorerHandler{
		client: client,
		coll:   coll,
		ctx:    ctx,
	}, nil
}

type mongoEntry struct {
	Key   string `bson:"_id"`
	Value []byte `bson:"value"`
}

func (r *mongodbStorerHandler) Put(key []byte, data []byte) error {
	// entry := &mongoEntry{
	// 	Key:   string(key),
	// 	Value: data,
	// }

	filter := bson.D{{"_id", string(key)}}
	update := bson.D{{"$set", bson.D{{"value", data}}}}

	opts := options.Update().SetUpsert(true)

	_, err := r.coll.UpdateOne(r.ctx, filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

func (r *mongodbStorerHandler) Get(key []byte) ([]byte, error) {
	filter := bson.D{{"_id", string(key)}}

	entry := &mongoEntry{}
	err := r.coll.FindOne(r.ctx, filter).Decode(entry)
	if err != nil {
		return nil, err
	}

	return entry.Value, nil
}

func (r *mongodbStorerHandler) Has(key []byte) error {
	filter := bson.D{{"_id", string(key)}}

	entry := &mongoEntry{}
	err := r.coll.FindOne(r.ctx, filter).Decode(entry)
	if err != nil {
		return err
	}

	return nil
}

func (r *mongodbStorerHandler) SearchFirst(key []byte) ([]byte, error) {
	return r.Get(key)
}

func (r *mongodbStorerHandler) Remove(key []byte) error {
	filter := bson.D{{"_id", string(key)}}
	_, err := r.coll.DeleteOne(r.ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

func (r *mongodbStorerHandler) ClearCache() {
	log.Warn("ClearCache: NOT implemented")
}

func (r *mongodbStorerHandler) Close() error {
	return r.client.Close(r.ctx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (r *mongodbStorerHandler) IsInterfaceNil() bool {
	return r == nil
}
