package mongodb

import (
	"context"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateMongoDBClient will create a new mongodb client
func CreateMongoDBClient(cfg config.MongoDBConfig) (*mongo.Client, error) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	return client, nil
}
