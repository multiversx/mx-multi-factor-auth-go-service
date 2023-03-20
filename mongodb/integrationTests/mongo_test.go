package integrationtests

import (
	"os"
	"sync"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tryvium-travels/memongo"
)

func TestMongoDBClient_ConcurrentCalls(t *testing.T) {
	t.Parallel()

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

	numCalls := 6

	var wg sync.WaitGroup
	wg.Add(numCalls)
	for i := 0; i < numCalls; i++ {
		go func(idx int) {
			switch idx % 5 {
			case 0:
				err := client.PutStruct(mongodb.UsersCollectionID, []byte("key"), &core.OTPInfo{LastTOTPChangeTimestamp: 101})
				require.Nil(t, err)
			case 1:
				_, err := client.GetStruct(mongodb.UsersCollectionID, []byte("key"))
				require.Nil(t, err)
			case 2:
				require.Nil(t, client.HasStruct(mongodb.UsersCollectionID, []byte("key")))
			case 3:
				err := client.ReadWriteWithCheck(mongodb.UsersCollectionID, []byte("key"), checker)
				require.Nil(t, err)
			case 4:
				_, err := client.UpdateTimestamp(mongodb.UsersCollectionID, []byte("key"), 0)
				require.Nil(t, err)
			default:
				assert.Fail(t, "should not hit default")
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
