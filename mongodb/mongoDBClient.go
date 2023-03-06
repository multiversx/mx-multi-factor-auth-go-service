package mongodb

import (
	"context"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongodbClient struct {
	client      *mongo.Client
	collections map[string]*mongo.Collection
}

// NewMongoDBClient will create a new mongodb client instance
func NewMongoDBClient(cfg config.MongoDBConfig) (*mongodbClient, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	collections := make(map[string]*mongo.Collection)
	collections["users"] = client.Database(cfg.DBName).Collection("users")

	return &mongodbClient{
		client:      client,
		collections: collections,
	}, nil
}

func (mdc *mongodbClient) GetCollection(coll string) *mongo.Collection {
	return mdc.collections[coll]
}

func (mdc *mongodbClient) Close(ctx context.Context) error {
	return mdc.client.Disconnect(ctx)
}

// IsInterfaceNil returns true if there is no value under the interface
func (mdc *mongodbClient) IsInterfaceNil() bool {
	return mdc == nil
}
