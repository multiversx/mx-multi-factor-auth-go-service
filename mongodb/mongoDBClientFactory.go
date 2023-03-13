package mongodb

import (
	"github.com/multiversx/multi-factor-auth-go-service/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateMongoDBClient will create a new mongo db client instance
func CreateMongoDBClient(cfg config.MongoDBConfig) (MongoDBClient, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(cfg.URI))
	if err != nil {
		return nil, err
	}

	clientWrapper := newMongoDBClientWrapper(client)

	return NewClient(clientWrapper, cfg.DBName)
}
