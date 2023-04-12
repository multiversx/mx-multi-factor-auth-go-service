package mongodb

import (
	"context"
	"fmt"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
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
)

const incrementIndexStep = 1
const minNumUsersColls = 1

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
func NewClient(client *mongo.Client, dbName string, numUsersColls uint32) (*mongodbClient, error) {
	if client == nil {
		return nil, ErrNilMongoDBClient
	}
	if dbName == "" {
		return nil, ErrEmptyMongoDBName
	}
	if numUsersColls < minNumUsersColls {
		return nil, fmt.Errorf("%w for number of users collections: provided %d, minimum %d",
			core.ErrInvalidValue, numUsersColls, minNumUsersColls)
	}

	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	database := client.Database(dbName)

	mongoClient := &mongodbClient{
		client: client,
		db:     database,
		ctx:    ctx,
	}

	mongoClient.createCollections(numUsersColls)

	return mongoClient, nil
}

func (mdc *mongodbClient) createCollections(numUsersColls uint32) {
	collections := make(map[CollectionID]*mongo.Collection)

	for i := uint32(0); i < numUsersColls; i++ {
		collName := fmt.Sprintf("%s_%d", string(UsersCollectionID), i)
		collections[CollectionID(collName)] = mdc.db.Collection(collName)
	}

	mdc.collections = collections
}

// GetAllCollectionsNames will returns collections names as array of strings
func (mdc *mongodbClient) GetAllCollectionsNames() []string {
	collectionNames := make([]string, 0)

	for _, coll := range mdc.collections {
		collectionNames = append(collectionNames, coll.Name())
	}

	return collectionNames
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
		if err.Error() == mongo.ErrNoDocuments.Error() {
			return nil, storage.ErrKeyNotFound
		}

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

// GetIndex will return the index value for the provided key and collection
func (mdc *mongodbClient) GetIndex(collID CollectionID, key []byte) (uint32, error) {
	coll, ok := mdc.collections[collID]
	if !ok {
		return 0, ErrCollectionNotFound
	}

	filter := bson.D{{Key: "_id", Value: string(key)}}

	entry := &counterMongoEntry{}
	err := coll.FindOne(mdc.ctx, filter).Decode(entry)
	if err != nil {
		return 0, err
	}

	return entry.Value, nil
}

// PutIndexIfNotExists will set an index value to the specified key if not already exists
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

	log.Trace("PutIndexIfNotExists", "collID", coll.Name(), "key", string(key), "value", index, "modifiedCount", res.ModifiedCount, "upsertedCount", res.UpsertedCount)

	return nil
}

// IncrementIndex will increment the value for the provided key
func (mdc *mongodbClient) IncrementIndex(collID CollectionID, key []byte) (uint32, error) {
	coll, ok := mdc.collections[collID]
	if !ok {
		return 0, ErrCollectionNotFound
	}

	opts := options.FindOneAndUpdate().SetUpsert(true)
	opts.SetReturnDocument(options.After)
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

	log.Trace("IncrementIndex", "collID", coll.Name(), "key", string(key), "value", entry.Value)

	return entry.Value, nil
}

// Close will close the mongodb client
func (mdc *mongodbClient) Close() error {
	err := mdc.client.Disconnect(mdc.ctx)
	if err == mongo.ErrClientDisconnected {
		log.Warn("MongoDBClient: client is already disconected")
		return nil
	}
	if err != nil {

		return err
	}

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (mdc *mongodbClient) IsInterfaceNil() bool {
	return mdc == nil
}
