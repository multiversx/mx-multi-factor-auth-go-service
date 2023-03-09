package factory

import (
	"fmt"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage/bucket"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/mx-chain-storage-go/storageUnit"
)

type shardedStorageFactory struct {
	cfg config.Config
}

// NewShardedStorageFactory returns a new instance of shardedStorageFactory
func NewShardedStorageFactory(config config.Config) *shardedStorageFactory {
	return &shardedStorageFactory{
		cfg: config,
	}
}

// Create returns a new instance of ShardedStorageWithIndex
func (ssf *shardedStorageFactory) Create() (core.ShardedStorageWithIndex, error) {
	switch ssf.cfg.ShardedStorage.DBType {
	case core.LevelDB:
		return ssf.createLocalDB()
	case core.MongoDB:
		return ssf.createLocalDB()
	default:
		// TODO: implement other types of storage
		return nil, handlers.ErrInvalidConfig
	}
}

func (ssf *shardedStorageFactory) createMongoDB() (core.ShardedStorageWithIndex, error) {
	client, err := mongodb.CreateMongoDBClient(ssf.cfg.MongoDB)
	if err != nil {
		return nil, err
	}

	storer, err := storage.NewMongoDBStorerHandler(client, mongodb.UsersCollectionID)
	if err != nil {
		return nil, err
	}

	bucketIDProvider, err := bucket.NewBucketIDProvider(1)
	if err != nil {
		return nil, err
	}

	bucketIndexHandlers := make(map[uint32]core.BucketIndexHandler, 1)
	bucketIndexHandlers[0], err = bucket.NewBucketIndexHandler(storer)
	if err != nil {
		return nil, err
	}

	argsShardedStorageWithIndex := bucket.ArgShardedStorageWithIndex{
		BucketIDProvider: bucketIDProvider,
		BucketHandlers:   bucketIndexHandlers,
	}

	return bucket.NewShardedStorageWithIndex(argsShardedStorageWithIndex)
}

func (ssf *shardedStorageFactory) createLocalDB() (core.ShardedStorageWithIndex, error) {
	numbOfBuckets := ssf.cfg.Buckets.NumberOfBuckets
	bucketIDProvider, err := bucket.NewBucketIDProvider(numbOfBuckets)
	if err != nil {
		return nil, err
	}

	localDBCfg := ssf.cfg.ShardedStorage.Users
	bucketIndexHandlers := make(map[uint32]core.BucketIndexHandler, numbOfBuckets)
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
func (ssf *shardedStorageFactory) IsInterfaceNil() bool {
	return ssf == nil
}
