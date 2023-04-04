package mongodb

import (
	"context"
	"fmt"

	logger "github.com/multiversx/mx-chain-logger-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var log = logger.GetOrCreate("mongodb")

// CollectionID defines mongodb collection type
type CollectionID string

const (
	// UsersCollectionID specifies mongodb collection for users
	UsersCollectionID CollectionID = "users"

	// IndexCollectionID specifies mongodb collection for global index
	IndexCollectionID CollectionID = "index"
)

const numInitialShardChunks = 4
const incrementIndexStep = 1

type mongoEntry struct {
	Key   string `bson:"_id"`
	Value []byte `bson:"value"`
}

type counterMongoEntry struct {
	Key   string `bson:"_id"`
	Value uint32 `bson:"value"`
}

type mongodbClient struct {
	client      *mongo.Client
	db          *mongo.Database
	collections map[CollectionID]*mongo.Collection
	ctx         context.Context
}

// NewClient will create a new mongodb client instance
func NewClient(client *mongo.Client, dbName string) (*mongodbClient, error) {
	if client == nil {
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

	database := client.Database(dbName)

	collections := make(map[CollectionID]*mongo.Collection)
	collections[UsersCollectionID] = database.Collection(string(UsersCollectionID))
	collections[IndexCollectionID] = database.Collection(string(IndexCollectionID))

	return &mongodbClient{
		client:      client,
		db:          database,
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

	log.Trace("Put", "key", string(key), "value", string(data))

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

	log.Debug("Get", "key", string(key))

	return entry.Value, nil
}

// Has will return true if the provided key exists in the collection
func (mdc *mongodbClient) Has(collID CollectionID, key []byte) error {
	_, err := mdc.findOne(collID, key)
	log.Debug("Has", "key", string(key))
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

func (mdc *mongodbClient) PutIndexIfNotExists(collID CollectionID, key []byte, index uint32) error {
	coll, ok := mdc.collections[collID]
	if !ok {
		return ErrCollectionNotFound
	}

	filter := bson.D{{Key: "_id", Value: string(key)}}
	update := bson.D{{Key: "$setOnInsert",
		Value: bson.D{
			{Key: "_id", Value: string(key)},
			{Key: "value", Value: index},
		},
	}}

	opts := options.Update().SetUpsert(true)

	res, err := coll.UpdateOne(mdc.ctx, filter, update, opts)
	if err != nil {
		return err
	}

	log.Trace("PutIndexIfNotExists", "key", string(key), "value", index, "modifiedCount", res.ModifiedCount)

	return nil
}

// IncrementIndex will increment the value for the provided key
func (mdc *mongodbClient) IncrementIndex(collID CollectionID, key []byte) (uint32, error) {
	coll, ok := mdc.collections[collID]
	if !ok {
		return 0, ErrCollectionNotFound
	}

	opts := options.FindOneAndUpdate().SetUpsert(true)
	filter := bson.D{{Key: "_id", Value: string(key)}}
	update := bson.D{{
		Key: "$inc",
		Value: bson.D{
			{Key: "value", Value: incrementIndexStep},
		},
	}}

	entry := &counterMongoEntry{}
	res := coll.FindOneAndUpdate(mdc.ctx, filter, update, opts)
	err := res.Decode(entry)
	if err != nil {
		return 0, err
	}

	log.Trace("IncrementIndex", "key", string(key), "value", entry.Value)

	return entry.Value, nil
}

// ShardHashedCollection will shard collection with a hashed shard key
func (mdc *mongodbClient) ShardHashedCollection(collID CollectionID) error {
	coll, ok := mdc.collections[collID]
	if !ok {
		return ErrCollectionNotFound
	}

	collectionPath := fmt.Sprintf("%s.%s", mdc.db.Name(), coll.Name())

	cmd := bson.D{
		{Key: "shardCollection", Value: collectionPath},
		{Key: "key", Value: bson.D{{Key: "_id", Value: "hashed"}}},
		{Key: "numInitialChunks", Value: numInitialShardChunks},
	}

	err := mdc.db.RunCommand(mdc.ctx, cmd).Err()
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
