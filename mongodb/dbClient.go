package mongodb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/mx-multi-factor-auth-go-service/metrics"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var log = logger.GetOrCreate("mongodb")

// CollectionID defines mongodb collection type
type CollectionID string

const (
	// UsersCollectionID specifies mongodb collection for users
	UsersCollectionID CollectionID = "users"

	metricPrefix        = "MongoDB"
	getIndexMetricLabel = "GetIndex"
	delMetricLabel      = "DeleteOne"
	findMetricLabel     = "FindOne"
	updateMetricLabel   = "UpdateOne"
	incMetricLabel      = "Increment"

	findPrimaryMetricLabel       = "FindOnePrimary"
	insertMetricLabel            = "InsertOne"
	conditionalUpdateMetricLabel = "UpdateOneConditional"
	versionConflictMetricLabel   = "VersionConflict"
)

const (
	initialVersion = 1
	conflictCode   = 409
)

const incrementIndexStep = 1
const minNumUsersColls = 1

type mongoEntry struct {
	Key     string `bson:"_id"`
	Value   []byte `bson:"value"`
	Version int64  `bson:"version,omitempty"` // missing means 0
}

type counterMongoEntry struct {
	Key   string `bson:"_id"`
	Value uint32 `bson:"value"`
}

type mongodbClient struct {
	client         *mongo.Client
	db             *mongo.Database
	collections    map[CollectionID]*mongo.Collection
	primaryColls   map[CollectionID]*mongo.Collection
	collectionsIDs []CollectionID
	ctx            context.Context
	metricsHandler core.StatusMetricsHandler
}

// NewClient will create a new mongodb client instance
func NewClient(client *mongo.Client, dbName string, numUsersColls uint32, metricsHandler core.StatusMetricsHandler) (*mongodbClient, error) {
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
	if check.IfNil(metricsHandler) {
		return nil, core.ErrNilMetricsHandler
	}

	ctx := context.Background()

	err := client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	database := client.Database(dbName)

	mongoClient := &mongodbClient{
		client:         client,
		db:             database,
		ctx:            ctx,
		metricsHandler: metricsHandler,
	}

	mongoClient.createCollections(numUsersColls)

	return mongoClient, nil
}

func (mdc *mongodbClient) createCollections(numUsersColls uint32) {
	collections := make(map[CollectionID]*mongo.Collection)
	primaryColls := make(map[CollectionID]*mongo.Collection)
	collectionIDs := make([]CollectionID, 0, len(mdc.collections))
	primaryOpts := options.Collection().SetReadPreference(readpref.Primary())

	for i := uint32(0); i < numUsersColls; i++ {
		collName := fmt.Sprintf("%s_%d", string(UsersCollectionID), i)
		collections[CollectionID(collName)] = mdc.db.Collection(collName)
		primaryColls[CollectionID(collName)] = mdc.db.Collection(collName, primaryOpts)
		collectionIDs = append(collectionIDs, CollectionID(collName))
	}

	mdc.collections = collections
	mdc.primaryColls = primaryColls
	mdc.collectionsIDs = collectionIDs
}

// GetAllCollectionsIDs returns collections names as array of collection ids
func (mdc *mongodbClient) GetAllCollectionsIDs() []CollectionID {
	return mdc.collectionsIDs
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

	t := time.Now()
	_, err := coll.UpdateOne(mdc.ctx, filter, update, opts)
	duration := time.Since(t)
	if err != nil {
		return err
	}
	mdc.metricsHandler.AddRequestData(getOpID(updateMetricLabel), duration, metrics.NonErrorCode)

	return nil
}

func (mdc *mongodbClient) findOne(collID CollectionID, key []byte) (*mongoEntry, error) {
	coll, ok := mdc.collections[collID]
	if !ok {
		return nil, ErrCollectionNotFound
	}

	filter := bson.D{{Key: "_id", Value: string(key)}}
	entry := &mongoEntry{}

	t := time.Now()
	err := coll.FindOne(mdc.ctx, filter).Decode(entry)
	duration := time.Since(t)
	if err != nil {
		return nil, err
	}
	mdc.metricsHandler.AddRequestData(getOpID(findMetricLabel), duration, metrics.NonErrorCode)

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

// GetWithVersion will return the value and its version for the provided key, reading from the primary
func (mdc *mongodbClient) GetWithVersion(collID CollectionID, key []byte) ([]byte, int64, error) {
	coll, ok := mdc.primaryColls[collID]
	if !ok {
		return nil, 0, ErrCollectionNotFound
	}

	filter := bson.D{{Key: "_id", Value: string(key)}}
	entry := &mongoEntry{}

	t := time.Now()
	err := coll.FindOne(mdc.ctx, filter).Decode(entry)
	duration := time.Since(t)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, 0, storage.ErrKeyNotFound
		}

		return nil, 0, err
	}
	mdc.metricsHandler.AddRequestData(getOpID(findPrimaryMetricLabel), duration, metrics.NonErrorCode)

	return entry.Value, entry.Version, nil
}

// Insert will add a new entry with the initial version, failing if the key already exists
func (mdc *mongodbClient) Insert(collID CollectionID, key []byte, data []byte) error {
	coll, ok := mdc.collections[collID]
	if !ok {
		return ErrCollectionNotFound
	}

	doc := bson.D{
		{Key: "_id", Value: string(key)},
		{Key: "value", Value: data},
		{Key: "version", Value: int64(initialVersion)},
	}

	t := time.Now()
	_, err := coll.InsertOne(mdc.ctx, doc)
	duration := time.Since(t)
	if mongo.IsDuplicateKeyError(err) {
		mdc.metricsHandler.AddRequestData(getOpID(versionConflictMetricLabel), duration, conflictCode)
		return storage.ErrKeyAlreadyExists
	}
	if err != nil {
		return err
	}
	mdc.metricsHandler.AddRequestData(getOpID(insertMetricLabel), duration, metrics.NonErrorCode)

	return nil
}

// PutIfVersion will update the entry only if its version matches the expected one, never creating it
func (mdc *mongodbClient) PutIfVersion(collID CollectionID, key []byte, data []byte, expectedVersion int64) error {
	coll, ok := mdc.collections[collID]
	if !ok {
		return ErrCollectionNotFound
	}

	filter := bson.D{{Key: "_id", Value: string(key)}}
	if expectedVersion == 0 {
		// matches missing, null or 0
		filter = append(filter, bson.E{Key: "version", Value: bson.D{{Key: "$in", Value: bson.A{nil, int64(0)}}}})
	} else {
		filter = append(filter, bson.E{Key: "version", Value: expectedVersion})
	}

	update := bson.D{
		{Key: "$set", Value: bson.D{{Key: "value", Value: data}}},
		{Key: "$inc", Value: bson.D{{Key: "version", Value: int64(1)}}},
	}

	t := time.Now()
	res, err := coll.UpdateOne(mdc.ctx, filter, update)
	duration := time.Since(t)
	if err != nil {
		return err
	}
	if res.MatchedCount == 0 {
		mdc.metricsHandler.AddRequestData(getOpID(versionConflictMetricLabel), duration, conflictCode)
		return storage.ErrVersionMismatch
	}
	mdc.metricsHandler.AddRequestData(getOpID(conditionalUpdateMetricLabel), duration, metrics.NonErrorCode)

	return nil
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

	t := time.Now()
	_, err := coll.DeleteOne(mdc.ctx, filter)
	duration := time.Since(t)
	if err != nil {
		return err
	}
	mdc.metricsHandler.AddRequestData(getOpID(delMetricLabel), duration, metrics.NonErrorCode)

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

	t := time.Now()
	err := coll.FindOne(mdc.ctx, filter).Decode(entry)
	duration := time.Since(t)
	if err != nil {
		return 0, err
	}
	mdc.metricsHandler.AddRequestData(getOpID(getIndexMetricLabel), duration, metrics.NonErrorCode)

	return entry.Value, nil
}

func getOpID(operation string) string {
	return fmt.Sprintf("%s-%s", metricPrefix, operation)
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

	t := time.Now()
	res := coll.FindOneAndUpdate(mdc.ctx, filter, update, opts)
	err := res.Decode(entry)
	duration := time.Since(t)
	if err != nil {
		return 0, err
	}
	mdc.metricsHandler.AddRequestData(getOpID(incMetricLabel), duration, metrics.NonErrorCode)

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

	return err
}

// IsInterfaceNil returns true if there is no value under the interface
func (mdc *mongodbClient) IsInterfaceNil() bool {
	return mdc == nil
}
