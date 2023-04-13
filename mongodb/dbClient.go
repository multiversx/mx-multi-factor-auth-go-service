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
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
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

const initialCounterValue = 1
const numInitialShardChunks = 4
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

type otpInfoWrapper struct {
	Key     string        `bson:"_id"`
	OTPInfo *core.OTPInfo `bson:"otpinfo"`
}

type mongodbClient struct {
	client         *mongo.Client
	db             *mongo.Database
	collections    map[CollectionID]*mongo.Collection
	collectionsIDs []CollectionID
	ctx            context.Context
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
	collectionIDs := make([]CollectionID, 0, len(mdc.collections))

	for i := uint32(0); i < numUsersColls; i++ {
		collName := fmt.Sprintf("%s_%d", string(UsersCollectionID), i)
		collections[CollectionID(collName)] = mdc.db.Collection(collName)
		collectionIDs = append(collectionIDs, CollectionID(collName))
	}

	mdc.collections = collections
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

	log.Trace("Put", "key", string(key), "value", string(data))

	_, err := coll.UpdateOne(mdc.ctx, filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

func (mdc *mongodbClient) PutStruct(collID CollectionID, key []byte, data *core.OTPInfo) error {
	coll, ok := mdc.collections[collID]
	if !ok {
		return ErrCollectionNotFound
	}

	otpInfo := &otpInfoWrapper{
		Key:     string(key),
		OTPInfo: data,
	}

	filter := bson.M{"_id": string(key)}
	update := bson.M{
		"$set": otpInfo,
	}

	opts := options.Update().SetUpsert(true)

	log.Trace("PutStruct", "key", string(key), "value", data.LastTOTPChangeTimestamp)

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

	log.Trace("Get", "key", string(key))

	return entry.Value, nil
}

func (mdc *mongodbClient) GetStruct(collID CollectionID, key []byte) (*core.OTPInfo, error) {
	coll, ok := mdc.collections[collID]
	if !ok {
		return nil, ErrCollectionNotFound
	}

	filter := bson.D{{Key: "_id", Value: string(key)}}

	entry := &otpInfoWrapper{}
	err := coll.FindOne(mdc.ctx, filter).Decode(entry)
	if err != nil {
		return nil, err
	}

	return entry.OTPInfo, nil
}

// Has will return true if the provided key exists in the collection
func (mdc *mongodbClient) Has(collID CollectionID, key []byte) error {
	_, err := mdc.findOne(collID, key)
	log.Trace("Has", "key", string(key))
	return err
}

func (mdc *mongodbClient) HasStruct(collID CollectionID, key []byte) error {
	coll, ok := mdc.collections[collID]
	if !ok {
		return ErrCollectionNotFound
	}

	filter := bson.D{{Key: "_id", Value: string(key)}}

	entry := &otpInfoWrapper{}
	err := coll.FindOne(mdc.ctx, filter).Decode(entry)
	if err != nil {
		return err
	}

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

// ReadWriteWithCheck will perform read and write operation with a provided checker
func (mdc *mongodbClient) ReadWriteWithCheck(
	collID CollectionID,
	key []byte,
	checker func(data interface{}) (interface{}, error),
) error {
	session, err := mdc.client.StartSession()
	if err != nil {
		return err
	}
	defer session.EndSession(mdc.ctx)

	wc := writeconcern.New(writeconcern.WMajority())
	txnOptions := options.Transaction().SetWriteConcern(wc)
	txnOptions.SetReadPreference(readpref.Primary())

	sessionCallback := func(ctx mongo.SessionContext) error {
		err := session.StartTransaction(txnOptions)
		if err != nil {
			return err
		}

		coll, ok := mdc.collections[collID]
		if !ok {
			return ErrCollectionNotFound
		}

		filter := bson.M{"_id": string(key)}

		entry := &otpInfoWrapper{}
		err = coll.FindOne(ctx, filter).Decode(entry)
		if err != nil {
			_ = session.AbortTransaction(ctx)
			return err
		}

		retValue, err := checker(entry.OTPInfo)
		if err != nil {
			_ = session.AbortTransaction(ctx)
			return err
		}
		retValueBytes, ok := retValue.(*core.OTPInfo)
		if !ok {
			_ = session.AbortTransaction(ctx)
			return core.ErrInvalidValue
		}

		otpInfo := &otpInfoWrapper{
			Key:     string(key),
			OTPInfo: retValueBytes,
		}

		update := bson.M{
			"$set": otpInfo,
		}

		opts := options.Update().SetUpsert(true)

		_, err = coll.UpdateOne(mdc.ctx, filter, update, opts)
		if err != nil {
			_ = session.AbortTransaction(ctx)
			return err
		}

		return session.CommitTransaction(ctx)
	}

	err = mongo.WithSession(mdc.ctx, session,
		func(sctx mongo.SessionContext) error {
			return runTxWithRetry(sctx, sessionCallback)
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func runTxWithRetry(sctx mongo.SessionContext, txnFn func(mongo.SessionContext) error) error {
	for {
		err := txnFn(sctx)
		if err == nil {
			return nil
		}

		log.Trace("Transaction aborted. Caught exception during transaction.")

		cmdErr, ok := err.(mongo.CommandError)
		if ok && cmdErr.HasErrorLabel("TransientTransactionError") {
			log.Trace("TransientTransactionError, retrying transaction...")
			continue
		}

		return err
	}
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
