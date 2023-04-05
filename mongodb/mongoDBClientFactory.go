package mongodb

import (
	"errors"
	"fmt"
	"time"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const minTimeoutInSec = 1

var errEmptyMongoURI = errors.New("empty mongo uri")

// CreateMongoDBClient will create a new mongo db client instance
func CreateMongoDBClient(cfg config.MongoDBConfig) (MongoDBClient, error) {
	err := checkMongoDBConfig(cfg)
	if err != nil {
		return nil, err
	}

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

func checkMongoDBConfig(cfg config.MongoDBConfig) error {
	if cfg.URI == "" {
		return errEmptyMongoURI
	}
	if cfg.ConnectTimeoutInSec < minTimeoutInSec {
		return fmt.Errorf("%w for mongo connect timeout: provided %d, minimum %d",
			core.ErrInvalidValue, cfg.ConnectTimeoutInSec, minTimeoutInSec)
	}
	if cfg.OperationTimeoutInSec < minTimeoutInSec {
		return fmt.Errorf("%w for mongo operation timeout: provided %d, minimum %d",
			core.ErrInvalidValue, cfg.OperationTimeoutInSec, minTimeoutInSec)
	}

	return nil
}
