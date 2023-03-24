package mongodb

import (
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CreateMongoDBClient will create a new mongo db client instance
func CreateMongoDBClient(cfg config.MongoDBConfig) (MongoDBClient, error) {
	opts := options.Client()
	opts.SetConnectTimeout(time.Duration(cfg.ConnectTimeoutInSec) * time.Second)
	opts.SetTimeout(time.Duration(cfg.OperationTimeoutInSec) * time.Second)
	opts.ApplyURI(cfg.URI)

	client, err := mongo.NewClient(opts)
	if err != nil {
		return nil, err
	}

	return NewClient(client, cfg.DBName)
}
