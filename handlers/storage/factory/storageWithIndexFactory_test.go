package factory

import (
	"fmt"
	"os"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/handlers"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-storage-go/storageUnit"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tryvium-travels/memongo"
)

func TestNewShardedStorageFactory_Create(t *testing.T) {
	t.Parallel()

	t.Run("should return error for flag unknown", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			ShardedStorage: config.ShardedStorageConfig{
				DBType: "dummy",
			},
		}
		ssf := NewStorageWithIndexFactory(cfg, config.ExternalConfig{})
		assert.NotNil(t, ssf)
		shardedStorageInstance, err := ssf.Create()
		assert.Equal(t, handlers.ErrInvalidConfig, err)
		assert.Nil(t, shardedStorageInstance)
	})
	t.Run("NewBucketIDProvider returns error", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			ShardedStorage: config.ShardedStorageConfig{
				DBType: core.LevelDB,
			},
			Buckets: config.BucketsConfig{
				NumberOfBuckets: 0,
			},
		}
		ssf := NewStorageWithIndexFactory(cfg, config.ExternalConfig{})
		assert.NotNil(t, ssf)
		shardedStorageInstance, err := ssf.Create()
		assert.NotNil(t, err)
		assert.Nil(t, shardedStorageInstance)
	})
	t.Run("NewStorageUnitFromConf returns error", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			ShardedStorage: config.ShardedStorageConfig{
				DBType: core.LevelDB,
				Users: config.StorageConfig{
					DB: storageUnit.DBConfig{
						MaxBatchSize: 100,
					},
					Cache: storageUnit.CacheConfig{
						Capacity: 10, // lower than DB.MaxBatchSize returns error
					},
				},
			},
			Buckets: config.BucketsConfig{
				NumberOfBuckets: 1,
			},
		}
		ssf := NewStorageWithIndexFactory(cfg, config.ExternalConfig{})
		assert.NotNil(t, ssf)
		shardedStorageInstance, err := ssf.Create()
		assert.NotNil(t, err)
		assert.Nil(t, shardedStorageInstance)
	})
	t.Run("should create local storage", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			ShardedStorage: config.ShardedStorageConfig{
				DBType: core.LevelDB,
				Users: config.StorageConfig{
					Cache: storageUnit.CacheConfig{
						Name:        "UsersCache",
						Type:        "SizeLRU",
						SizeInBytes: 104857600,
						Capacity:    100000,
					},
					DB: storageUnit.DBConfig{
						FilePath:          "UsersDB",
						Type:              "LvlDB",
						BatchDelaySeconds: 1,
						MaxBatchSize:      1000,
						MaxOpenFiles:      10,
					},
				},
			},
			Buckets: config.BucketsConfig{
				NumberOfBuckets: 4,
			},
		}
		extCfg := config.ExternalConfig{
			Api: config.ApiConfig{
				NetworkAddress: "http://localhost:8080",
			},
			MongoDB: config.MongoDBConfig{
				DBName: "dbName",
			},
		}
		ssf := NewStorageWithIndexFactory(cfg, extCfg)
		assert.NotNil(t, ssf)
		shardedStorageInstance, err := ssf.Create()
		assert.Nil(t, err)
		assert.NotNil(t, shardedStorageInstance)
		assert.Equal(t, "*bucket.shardedStorageWithIndex", fmt.Sprintf("%T", shardedStorageInstance))
		removeDBs(t, cfg)
	})

	t.Run("real storage LevelDB, returns ErrKeyNotFound on non existing key", func(t *testing.T) {
		t.Parallel()

		cfg := config.Config{
			ShardedStorage: config.ShardedStorageConfig{
				DBType: core.LevelDB,
				Users: config.StorageConfig{
					Cache: storageUnit.CacheConfig{
						Name:        "UsersCache",
						Type:        "SizeLRU",
						SizeInBytes: 104857600,
						Capacity:    100000,
					},
					DB: storageUnit.DBConfig{
						FilePath:          "UsersDB",
						Type:              "LvlDB",
						BatchDelaySeconds: 1,
						MaxBatchSize:      1000,
						MaxOpenFiles:      10,
					},
				},
			},
			Buckets: config.BucketsConfig{
				NumberOfBuckets: 4,
			},
		}
		extCfg := config.ExternalConfig{
			Api: config.ApiConfig{
				NetworkAddress: "http://localhost:8080",
			},
			MongoDB: config.MongoDBConfig{
				DBName: "dbName",
			},
		}
		ssf := NewStorageWithIndexFactory(cfg, extCfg)
		assert.False(t, check.IfNil(ssf))
		shardedStorageInstance, err := ssf.Create()
		assert.Nil(t, err)
		assert.False(t, check.IfNil(shardedStorageInstance))

		_, err = shardedStorageInstance.Get([]byte("key"))
		assert.Equal(t, storage.ErrKeyNotFound, err)
		removeDBs(t, cfg)
	})

	t.Run("real storage MongoDB, returns ErrKeyNotFound on non existing key", func(t *testing.T) {
		t.Parallel()

		if os.Getenv("CI") != "" {
			t.Skip("Skipping testing in CI environment")
		}

		inMemoryMongoDB, err := memongo.StartWithOptions(&memongo.Options{MongoVersion: "4.4.0", ShouldUseReplica: true})
		require.Nil(t, err)
		defer inMemoryMongoDB.Stop()

		cfg := config.Config{
			ShardedStorage: config.ShardedStorageConfig{
				DBType: core.MongoDB,
			},
			Buckets: config.BucketsConfig{
				NumberOfBuckets: 1,
			},
		}
		extCfg := config.ExternalConfig{
			Api: config.ApiConfig{
				NetworkAddress: "http://localhost:8080",
			},
			MongoDB: config.MongoDBConfig{
				URI:                   inMemoryMongoDB.URI(),
				DBName:                "dbName",
				ConnectTimeoutInSec:   10,
				OperationTimeoutInSec: 10,
			},
		}

		ssf := NewStorageWithIndexFactory(cfg, extCfg)
		assert.False(t, check.IfNil(ssf))
		shardedStorageInstance, err := ssf.Create()
		assert.Nil(t, err)
		assert.False(t, check.IfNil(shardedStorageInstance))

		_, err = shardedStorageInstance.Get([]byte("key"))
		assert.Equal(t, storage.ErrKeyNotFound, err)
	})
}

func removeDBs(t *testing.T, cfg config.Config) {
	for i := uint32(0); i < cfg.Buckets.NumberOfBuckets; i++ {
		dirName := fmt.Sprintf("%s_%d", cfg.ShardedStorage.Users.DB.FilePath, i)
		assert.Nil(t, os.RemoveAll(dirName))
	}
}
