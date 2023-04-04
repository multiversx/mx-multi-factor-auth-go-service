package mongodb

import (
	"context"
	"fmt"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
	"go.mongodb.org/mongo-driver/x/mongo/driver"
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

var withTransactionTimeout = 10 * time.Second

type mongoEntry struct {
	Key   string `bson:"_id"`
	Value []byte `bson:"value"`
}

type counterMongoEntry struct {
	Key   string `bson:"_id"`
	Value uint32 `bson:"value"`
}

// TODO: change with merged structures
type otpInfoWrapper struct {
	Key string `bson:"_id"`
	*core.OTPInfo
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

func (mdc *mongodbClient) UpdateTimestamp(collID CollectionID, key []byte, interval int64) (int64, error) {
	coll, ok := mdc.collections[collID]
	if !ok {
		return 0, ErrCollectionNotFound
	}

	opts := options.FindOneAndUpdate().SetUpsert(false)

	currentTimestamp := time.Now().Unix()
	compareValue := currentTimestamp - interval

	filter := bson.M{"_id": string(key), "otpinfo.lasttotpchangetimestamp": bson.M{"$lt": compareValue}}
	update := bson.M{
		"$set": bson.M{
			"otpinfo.lasttotpchangetimestamp": time.Now().Unix(),
		},
	}

	entry := &core.OTPInfo{}
	res := coll.FindOneAndUpdate(mdc.ctx, filter, update, opts)
	err := res.Decode(entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return currentTimestamp, nil
		}

		return currentTimestamp, err
	}

	return entry.LastTOTPChangeTimestamp, nil
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

func (mdc *mongodbClient) ReadWithTx(
	collID CollectionID,
	key []byte,
) ([]byte, Session, SessionContext, error) {
	coll, ok := mdc.collections[collID]
	if !ok {
		return nil, nil, nil, ErrCollectionNotFound
	}

	session, err := mdc.client.StartSession()
	if err != nil {
		return nil, nil, nil, err
	}

	log.Trace("started session", "ID", session.ID())

	sessionCtx := mongo.NewSessionContext(mdc.ctx, session)

	wc := writeconcern.New(writeconcern.WMajority())
	txnOptions := options.Transaction().SetWriteConcern(wc)
	txnOptions.SetReadPreference(readpref.Primary())

	err = session.StartTransaction(txnOptions)
	if err != nil {
		log.Trace("ReadWithTx: StartTransaction", "err", err.Error())
		return nil, nil, nil, err
	}

	filter := bson.M{"_id": string(key)}

	entry := &mongoEntry{}
	err = coll.FindOne(sessionCtx, filter).Decode(entry)
	if err != nil {
		// TODO: abort transaction and create a new one on write
		//_ = session.AbortTransaction(sessionCtx)
		log.Trace("ReadWithTx", "err", err.Error())
		return nil, session, sessionCtx, storage.ErrKeyNotFound
	}

	log.Trace("ReadWithTx", "key", string(key), "value", entry.Value)

	return entry.Value, session, sessionCtx, nil
}

func (mdc *mongodbClient) WriteWithTx(
	collID CollectionID,
	key []byte,
	value []byte,
	session Session,
	sessionCtx SessionContext,
) error {

	txCallback := func(ctx mongo.SessionContext) error {
		coll, ok := mdc.collections[collID]
		if !ok {
			return ErrCollectionNotFound
		}

		filter := bson.M{"_id": string(key)}
		update := bson.M{
			"$set": bson.M{
				"_id":   string(key),
				"value": value,
			},
		}

		// filter := bson.D{{Key: "_id", Value: string(key)}}
		// update := bson.D{{Key: "$set",
		// 	Value: bson.D{
		// 		{Key: "_id", Value: string(key)},
		// 		{Key: "value", Value: value},
		// 	},
		// }}

		opts := options.Update().SetUpsert(true)

		_, err := coll.UpdateOne(sessionCtx, filter, update, opts)
		if err != nil {
			log.Trace("WriteWithTx: UpdateOne", "err", err.Error())
			//_ = session.AbortTransaction(sessionCtx)
			return err
		}

		log.Trace("WriteWithTx before commit", "key", string(key), "value", value)

		err = session.CommitTransaction(sessionCtx)
		if err != nil {
			log.Trace("WriteWithTx: CommitTransaction", "err", err.Error())
			//_ = session.AbortTransaction(mdc.ctx)
			return err
		}

		log.Trace("WriteWithTx", "key", string(key), "value", value)

		return nil
	}

	err := txCallback(sessionCtx)
	if err != nil {
		abortErr := session.AbortTransaction(mdc.ctx)
		if abortErr != nil {
			return abortErr
		}

		log.Trace("ended session", "ID", session.ID())
		session.EndSession(mdc.ctx)
		return err
	}

	log.Trace("ended session", "ID", session.ID())
	session.EndSession(mdc.ctx)

	return nil
}

func runTxWithRetry(sctx mongo.SessionContext, txnFn func(mongo.SessionContext) error) error {
	timeout := time.NewTimer(withTransactionTimeout)
	defer timeout.Stop()

	for {
		err := txnFn(sctx)
		if err == nil {
			return nil
		}

		time.Sleep(2 * time.Second)

		log.Trace("Transaction aborted. Caught exception during transaction.")

		select {
		case <-timeout.C:
			log.Trace("Transaction timeout reached.")
			return err
		default:
		}

		cmdErr, ok := err.(mongo.CommandError)
		if ok && cmdErr.HasErrorLabel(driver.TransientTransactionError) {
			log.Trace("TransientTransactionError, retrying transaction...")
			continue
		}

		log.Trace("other transaction error: %s", err.Error())
		return err
	}
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

	log.Trace("PutIfNotExists", "key", string(key), "value", index, "modifiedCount", res.ModifiedCount)

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
			{Key: "value", Value: uint32(1)},
		},
	}}

	entry := &counterMongoEntry{}
	res := coll.FindOneAndUpdate(mdc.ctx, filter, update, opts)
	err := res.Decode(entry)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			log.Trace(
				"IncrementIndex: no document found, will return initial counter value",
				"key", string(key),
				"value", entry.Value,
			)
			return initialCounterValue, nil
		}

		return initialCounterValue, err
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
