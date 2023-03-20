package mongodb

import (
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"
)

const (
	connectTimeoutSec   = 60
	operationTimeoutSec = 60
)

// CreateMongoDBClient will create a new mongo db client instance
func CreateMongoDBClient(cfg config.MongoDBConfig) (MongoDBUsersHandler, error) {
	opts := options.Client()
	opts.SetConnectTimeout(connectTimeoutSec * time.Second)
	opts.SetTimeout(operationTimeoutSec * time.Second)
	opts.ApplyURI(cfg.URI)

	writeConcern := writeconcern.New(writeconcern.WMajority())
	opts.SetWriteConcern(writeConcern)

	readPref, err := readpref.New(readpref.SecondaryPreferredMode)
	if err != nil {
		return nil, err
	}

	opts.SetReadPreference(readPref)

	client, err := mongo.NewClient(opts)
	if err != nil {
		return nil, err
	}

	return NewClient(client, cfg.DBName)
}
