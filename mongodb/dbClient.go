package mongodb

import (
	"context"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/core"
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
)

const initialCounterValue = 1

type mongoEntry struct {
	Key   string `bson:"_id"`
	Value []byte `bson:"value"`
}

type counterMongoEntry struct {
	Key   string `bson:"_id"`
	Value uint32 `bson:"value"`
}

type otpInfoWrapper struct {
	Key string `bson:"_id"`
	*core.OTPInfo
}

type mongodbClient struct {
	client      *mongo.Client
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

	collections := make(map[CollectionID]*mongo.Collection)
	collections[UsersCollectionID] = client.Database(dbName).Collection(string(UsersCollectionID))

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

// IncrementWithTransaction will increment the value for the provided key, within a transaction
func (mdc *mongodbClient) IncrementWithTransaction(collID CollectionID, key []byte) (uint32, error) {
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
			return initialCounterValue, nil
		}

		return initialCounterValue, err
	}

	return entry.Value, nil
}

// Close will close the mongodb client
func (mdc *mongodbClient) Close() error {
	return mdc.client.Disconnect(mdc.ctx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mdc *mongodbClient) IsInterfaceNil() bool {
	return mdc == nil
}
