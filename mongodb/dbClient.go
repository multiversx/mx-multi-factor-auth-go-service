package mongodb

import (
	"context"
	"encoding/binary"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	logger "github.com/multiversx/mx-chain-logger-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

var log = logger.GetOrCreate("mongodb")

// CollectionID defines mongodb collection type
type CollectionID string

const (
	// UsersCollectionID specifies mongodb collection for users
	UsersCollectionID CollectionID = "users"
)

type mongoEntry struct {
	Key   string `bson:"_id"`
	Value []byte `bson:"value"`
}

type mongodbClient struct {
	client      *mongo.Client
	collections map[CollectionID]MongoDBCollection
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

	collections := make(map[CollectionID]MongoDBCollection)
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

// Has will return true if the provided key exists in the collection
func (mdc *mongodbClient) Has(collID CollectionID, key []byte) error {
	_, err := mdc.findOne(collID, key)
	log.Trace("Has", "key", string(key))
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

	sessionCallback := func(ctx mongo.SessionContext) error {
		err := session.StartTransaction(txnOptions)
		if err != nil {
			return err
		}

		value, err := mdc.Get(collID, key)
		if err != nil {
			return err
		}

		retValue, err := checker(value)
		if err != nil {
			return err
		}
		retValueBytes, ok := retValue.([]byte)
		if !ok {
			return core.ErrInvalidValue
		}

		err = mdc.Put(collID, key, retValueBytes)

		if err = session.CommitTransaction(ctx); err != nil {
			return err
		}

		return nil
	}

	err = mongo.WithSession(mdc.ctx, session, sessionCallback)
	if err != nil {
		if err := session.AbortTransaction(mdc.ctx); err != nil {
			return err
		}
		return err
	}

	return nil
}

// IncrementWithTransaction will increment the value for the provided key, within a transaction
func (mdc *mongodbClient) IncrementWithTransaction(collID CollectionID, key []byte) (uint32, error) {
	coll, ok := mdc.collections[collID]
	if !ok {
		return 0, ErrCollectionNotFound
	}

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		filter := bson.D{{Key: "_id", Value: string(key)}}

		entry := &mongoEntry{}
		err := coll.FindOne(sessCtx, filter).Decode(entry)
		if err != nil {
			return nil, err
		}

		latestIndexBytes, newIndex := incrementIntegerFromBytes(entry.Value)

		filter = bson.D{{Key: "_id", Value: string(key)}}
		update := bson.D{{Key: "$set",
			Value: bson.D{
				{Key: "_id", Value: string(key)},
				{Key: "value", Value: latestIndexBytes},
			},
		}}

		opts := options.Update().SetUpsert(true)

		_, err = coll.UpdateOne(sessCtx, filter, update, opts)
		if err != nil {
			return nil, err
		}

		return newIndex, nil
	}

	session, err := mdc.client.StartSession()
	if err != nil {
		return 0, err
	}
	defer session.EndSession(mdc.ctx)

	newIndex, err := session.WithTransaction(mdc.ctx, callback)
	if err != nil {
		return 0, err
	}
	index, ok := newIndex.(uint32)
	if !ok {
		return 0, core.ErrInvalidValue
	}

	return index, nil
}

func incrementIntegerFromBytes(value []byte) ([]byte, uint32) {
	uint32Value := binary.BigEndian.Uint32(value)
	newIndex := uint32Value + 1

	newIndexBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(newIndexBytes, newIndex)

	return newIndexBytes, newIndex
}

// Close will close the mongodb client
func (mdc *mongodbClient) Close() error {
	return mdc.client.Disconnect(mdc.ctx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mdc *mongodbClient) IsInterfaceNil() bool {
	return mdc == nil
}
