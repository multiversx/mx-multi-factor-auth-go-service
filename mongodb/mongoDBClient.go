package mongodb

import (
	"context"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collection defines mongodb collection type
type Collection string

const (
	// UsersCollection specifies mongodb collection for users
	UsersCollection Collection = "users"
)

type mongoEntry struct {
	Key   string `bson:"_id"`
	Value []byte `bson:"value"`
}

type mongodbClient struct {
	client      *mongo.Client
	collections map[Collection]*mongo.Collection
	ctx         context.Context
}

// NewMongoDBClient will create a new mongodb client instance
func NewMongoDBClient(cfg config.MongoDBConfig) (*mongodbClient, error) {
	ctx := context.Background()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	collections := make(map[Collection]*mongo.Collection)
	collections[UsersCollection] = client.Database(cfg.DBName).Collection(string(UsersCollection))

	return &mongodbClient{
		client:      client,
		collections: collections,
		ctx:         ctx,
	}, nil
}

func (mdc *mongodbClient) GetCollection(coll Collection) *mongo.Collection {
	return mdc.collections[coll]
}

func (mdc *mongodbClient) Put(coll Collection, key []byte, data []byte) error {
	filter := bson.D{{Key: string(key)}}
	update := bson.D{primitive.E{Key: "$set",
		Value: bson.D{
			{Value: data},
		},
	}}

	opts := options.Update().SetUpsert(true)

	_, err := mdc.collections[coll].UpdateOne(mdc.ctx, filter, update, opts)
	if err != nil {
		return err
	}

	return nil
}

func (mdc *mongodbClient) Get(coll Collection, key []byte) ([]byte, error) {
	filter := bson.D{{Key: string(key)}}

	entry := &mongoEntry{}
	err := mdc.collections[coll].FindOne(mdc.ctx, filter).Decode(entry)
	if err != nil {
		return nil, err
	}

	return entry.Value, nil
}

func (mdc *mongodbClient) Has(coll Collection, key []byte) error {
	filter := bson.D{{Key: string(key)}}

	entry := &mongoEntry{}
	return mdc.collections[coll].FindOne(mdc.ctx, filter).Decode(entry)
}

func (mdc *mongodbClient) Remove(coll Collection, key []byte) error {
	filter := bson.D{{Key: string(key)}}
	_, err := mdc.collections[coll].DeleteOne(mdc.ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

func (mdc *mongodbClient) Close() error {
	return mdc.client.Disconnect(mdc.ctx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mdc *mongodbClient) IsInterfaceNil() bool {
	return mdc == nil
}
