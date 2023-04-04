package integrationtests

import (
	"os"
	"sync"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tryvium-travels/memongo"
)

func TestMongoDBClient_ConcurrentCalls(t *testing.T) {
	t.Parallel()

	logger.SetLogLevel("*:TRACE")

	if os.Getenv("CI") != "" {
		t.Skip("Skipping testing in CI environment")
	}

	inMemoryMongoDB, err := memongo.StartWithOptions(&memongo.Options{MongoVersion: "4.4.0", ShouldUseReplica: true})
	require.Nil(t, err)
	defer inMemoryMongoDB.Stop()

	client, err := mongodb.CreateMongoDBClient(config.MongoDBConfig{
		URI:    inMemoryMongoDB.URI(),
		DBName: memongo.RandomDatabase(),
	})
	require.Nil(t, err)

	checker := func(data interface{}) (interface{}, error) {
		return &core.OTPInfo{}, nil
	}

	err = client.PutStruct(mongodb.UsersCollectionID, []byte("key"), &core.OTPInfo{LastTOTPChangeTimestamp: 101})
	require.Nil(t, err)

	numCalls := 20

	var wg sync.WaitGroup
	wg.Add(numCalls)
	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			switch idx % 6 {
			case 0:
				err := client.PutStruct(mongodb.UsersCollectionID, []byte("key"), &core.OTPInfo{LastTOTPChangeTimestamp: 101})
				assert.Nil(t, err)
			case 1:
				_, err := client.GetStruct(mongodb.UsersCollectionID, []byte("key"))
				assert.Nil(t, err)
			case 2:
				assert.Nil(t, client.HasStruct(mongodb.UsersCollectionID, []byte("key")))
			case 3:
				err := client.ReadWriteWithCheck(mongodb.UsersCollectionID, []byte("key"), checker)
				assert.Nil(t, err)
			case 4:
				_, err := client.UpdateTimestamp(mongodb.UsersCollectionID, []byte("key"), 0)
				assert.Nil(t, err)
			case 5:
				_ = client.Put(mongodb.UsersCollectionID, []byte("key2"), []byte{1, 2, 3})
				_, sess, sessCtx, err := client.ReadWithTx(mongodb.UsersCollectionID, []byte("key2"))
				assert.Nil(t, err)

				err = client.WriteWithTx(mongodb.UsersCollectionID, []byte("key"), []byte("data"), sess, sessCtx)
				assert.Nil(t, err)
			default:
				assert.Fail(t, "should not hit default")
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
