package mongodb

import (
	"context"
	"errors"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CollectionID defines mongodb collection type
type CollectionID string

const (
	// UsersCollectionID specifies mongodb collection for users
	UsersCollectionID CollectionID = "users"
)

// ErrNilMongoDBClient signals that a nil mongodb client has been provided
var ErrNilMongoDBClient = errors.New("nil mongodb client")

type mongoEntry struct {
	Key   string `bson:"_id"`
	Value []byte `bson:"value"`
}

type mongodbClient struct {
	client      MongoDBClientWrapper
	collections map[CollectionID]MongoDBCollection
	ctx         context.Context
}

// NewClient will create a new mongodb client instance
func NewClient(client MongoDBClientWrapper, dbName string) (*mongodbClient, error) {
	if check.IfNil(client) {
		return nil, ErrNilMongoDBClientWrapper
	}
	if dbName == "" {
		return nil, ErrEmptyMongoDBName
	}

	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	collections := make(map[CollectionID]MongoDBCollection)
	collections[UsersCollectionID] = client.DBCollection(dbName, string(UsersCollectionID))

	return &mongodbClient{
		client:      client,
		collections: collections,
		ctx:         ctx,
	}, nil
}

// Put will set key value pair into specified collection
func (mdc *mongodbClient) Put(collID CollectionID, key []byte, data []byte) error {
	coll, ok := mdc.collections[collID]
	if !ok {
		return ErrCollectionNotFound
	}

	filter := bson.D{{Key: "_id", Value: string(key)}}
	update := bson.D{{Key: "$set",
		Value: bson.D{
			{Key: "_id", Value: string(key)},
			{Key: "value", Value: data},
		},
	}}

	opts := options.Update().SetUpsert(true)

	_, err := coll.UpdateOne(mdc.ctx, filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

func (mdc *mongodbClient) findOne(collID CollectionID, key []byte) (*mongoEntry, error) {
	coll, ok := mdc.collections[collID]
	if !ok {
		return nil, ErrCollectionNotFound
	}

	filter := bson.D{{Key: "_id", Value: string(key)}}

	entry := &mongoEntry{}
	err := coll.FindOne(mdc.ctx, filter).Decode(entry)
	if err != nil {
		return nil, err
	}

	return entry, nil
}

// Get will return the value for the provided key and collection
func (mdc *mongodbClient) Get(collID CollectionID, key []byte) ([]byte, error) {
	entry, err := mdc.findOne(collID, key)
	if err != nil {
		return nil, err
	}

	return entry.Value, nil
}

// Has will return true if the provided key exists in the collection
func (mdc *mongodbClient) Has(collID CollectionID, key []byte) error {
	_, err := mdc.findOne(collID, key)
	return err
}

// Remove will remove the provided key from the collection
func (mdc *mongodbClient) Remove(collID CollectionID, key []byte) error {
	coll, ok := mdc.collections[collID]
	if !ok {
		return ErrCollectionNotFound
	}

	filter := bson.D{{Key: "_id", Value: string(key)}}

	_, err := coll.DeleteOne(mdc.ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

// Close will close the mongodb client
func (mdc *mongodbClient) Close() error {
	return mdc.client.Disconnect(mdc.ctx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mdc *mongodbClient) IsInterfaceNil() bool {
	return mdc == nil
}
