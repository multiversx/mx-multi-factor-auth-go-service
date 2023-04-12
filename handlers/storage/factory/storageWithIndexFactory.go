package factory

import (
	"fmt"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage/bucket"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/mx-chain-storage-go/storageUnit"
)

const numBucketsForMongoStorage = 1

type storageWithIndexFactory struct {
	cfg         config.Config
	externalCfg config.ExternalConfig
}

// NewStorageWithIndexFactory returns a new instance of storageWithIndexFactory
func NewStorageWithIndexFactory(config config.Config, externalCfg config.ExternalConfig) *storageWithIndexFactory {
	return &storageWithIndexFactory{
		cfg:         config,
		externalCfg: externalCfg,
	}
}

// Create returns a new instance of StorageWithIndex component
func (ssf *storageWithIndexFactory) Create() (core.StorageWithIndex, error) {
	switch ssf.cfg.General.DBType {
	case core.LevelDB:
		return ssf.createLocalDB()
	case core.MongoDB:
		return ssf.createMongoDB()
	default:
		return nil, handlers.ErrInvalidConfig
	}
}

func (ssf *storageWithIndexFactory) createMongoDB() (core.StorageWithIndex, error) {
	client, err := mongodb.CreateMongoDBClient(ssf.externalCfg.MongoDB)
	if err != nil {
		return nil, err
	}

	numOfBuckets := ssf.externalCfg.MongoDB.NumUsersCollections

	bucketIDProvider, err := bucket.NewBucketIDProvider(numOfBuckets)
	if err != nil {
		return nil, err
	}

	indexHandlers := make(map[uint32]core.IndexHandler, numOfBuckets)
	for i, collName := range client.GetAllCollectionsNames() {
		indexHandlers[uint32(i)], err = bucket.NewMongoDBIndexHandler(client, collName)
		if err != nil {
			return nil, err
		}
	}

	argsShardedStorageWithIndex := bucket.ArgShardedStorageWithIndex{
		BucketIDProvider: bucketIDProvider,
		BucketHandlers:   indexHandlers,
	}

	return bucket.NewShardedStorageWithIndex(argsShardedStorageWithIndex)
}

func (ssf *storageWithIndexFactory) createLocalDB() (core.StorageWithIndex, error) {
	numbOfBuckets := ssf.cfg.Buckets.NumberOfBuckets
	bucketIDProvider, err := bucket.NewBucketIDProvider(numbOfBuckets)
	if err != nil {
		return nil, err
	}

	localDBCfg := ssf.cfg.ShardedStorage.Users
	bucketIndexHandlers := make(map[uint32]core.IndexHandler, numbOfBuckets)
	var bucketStorer core.Storer
	for i := uint32(0); i < numbOfBuckets; i++ {
		cacheCfg := localDBCfg.Cache
		cacheCfg.Name = fmt.Sprintf("%s_%d", cacheCfg.Name, i)
		dbCfg := localDBCfg.DB
		dbCfg.FilePath = fmt.Sprintf("%s_%d", dbCfg.FilePath, i)

		bucketStorer, err = storageUnit.NewStorageUnitFromConf(cacheCfg, dbCfg)
		if err != nil {
			return nil, err
		}

		bucketIndexHandlers[i], err = bucket.NewBucketIndexHandler(bucketStorer)
		if err != nil {
			return nil, err
		}
	}

	argsShardedStorageWithIndex := bucket.ArgShardedStorageWithIndex{
		BucketIDProvider: bucketIDProvider,
		BucketHandlers:   bucketIndexHandlers,
	}

	return bucket.NewShardedStorageWithIndex(argsShardedStorageWithIndex)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ssf *storageWithIndexFactory) IsInterfaceNil() bool {
	return ssf == nil
}
