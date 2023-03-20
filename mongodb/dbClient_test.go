package mongodb_test

import (
	"errors"
	"sync"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestNewMongoDBClient(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("nil client wrapper, should fail", func(mt *mtest.T) {
		client, err := mongodb.NewClient(nil, "dbName")
		require.Nil(mt, client)
		require.Equal(mt, mongodb.ErrNilMongoDBClientWrapper, err)
	})

	mt.Run("empty db name, should fail", func(mt *mtest.T) {
		client, err := mongodb.NewClient(mt.Client, "")
		require.Nil(mt, client)
		require.Equal(mt, mongodb.ErrEmptyMongoDBName, err)
	})

	mt.Run("should work", func(mt *mtest.T) {
		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)
		require.False(mt, client.IsInterfaceNil())
	})
}

func TestMongoDBClient_Put(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		err = client.Put("another coll", []byte("key1"), []byte("data"))
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
	})

	mt.Run("should fail", func(mt *mtest.T) {
		mt.Parallel()

		expectedErr := errors.New("expected error")
		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: expectedErr.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		err = client.Put(mongodb.UsersCollectionID, []byte("key1"), []byte("data"))
		require.Equal(mt, expectedErr.Error(), err.Error())
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key2"},
				{Key: "value", Value: []byte("value")},
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		err = client.Put(mongodb.UsersCollectionID, []byte("key1"), []byte("data"))
		require.Nil(mt, err)
	})
}

func TestMongoDBClient_Get(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		_, err = client.Get("another coll", []byte("key1"))
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
	})

	mt.Run("find one entry failed", func(mt *mtest.T) {
		mt.Parallel()

		expectedErr := errors.New("expected err")
		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: expectedErr.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		_, err = client.Get(mongodb.UsersCollectionID, []byte("key1"))
		require.Equal(mt, expectedErr.Error(), err.Error())
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key2"},
				{Key: "value", Value: []byte("value")},
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		_, err = client.Get(mongodb.UsersCollectionID, []byte("key1"))
		require.Nil(mt, err)
	})
}

func TestMongoDBClient_Has(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("should fail", func(mt *mtest.T) {
		mt.Parallel()

		expectedErr := errors.New("expected err")
		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: expectedErr.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		err = client.Has(mongodb.UsersCollectionID, []byte("key1"))
		require.Equal(mt, expectedErr.Error(), err.Error())
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key2"},
				{Key: "value", Value: []byte("value")},
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		err = client.Has(mongodb.UsersCollectionID, []byte("key1"))
		require.Nil(mt, err)
	})
}

func TestMongoDBClient_Remove(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		err = client.Remove("another coll", []byte("key1"))
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key1"},
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		err = client.Remove(mongodb.UsersCollectionID, []byte("key1"))
		require.Nil(mt, err)
	})
}

func TestMongoDBClient_IncrementWithTransaction(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("failed to create session", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		expectedErr := errors.New("expected error")

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: expectedErr.Error(),
			}),
		)

		_, err = client.IncrementWithTransaction(mongodb.UsersCollectionID, []byte("key1"))
		require.Equal(mt, expectedErr.Error(), err.Error())
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key1"},
				{Key: "value", Value: 1},
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		val, err := client.IncrementWithTransaction(mongodb.UsersCollectionID, []byte("key1"))
		require.Nil(mt, err)
		require.Equal(mt, uint32(1), val)
	})
}

func TestMongoDBClient_ReadWriteWithCheck(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("failed to create session", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		expectedErr := errors.New("expected error")

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: expectedErr.Error(),
			}),
		)

		checker := func(data interface{}) (interface{}, error) {
			return nil, nil
		}

		err = client.ReadWriteWithCheck(mongodb.UsersCollectionID, []byte("key1"), checker)
		require.Equal(mt, expectedErr.Error(), err.Error())
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key1"},
				{Key: "otpinfo", Value: &core.OTPInfo{LastTOTPChangeTimestamp: 101}},
			}),
			mtest.CreateSuccessResponse(),
			mtest.CreateSuccessResponse(),
			mtest.CreateSuccessResponse(),
		)

		checker := func(data interface{}) (interface{}, error) {
			if data.(*core.OTPInfo).LastTOTPChangeTimestamp == 101 {
				return &core.OTPInfo{}, nil
			}
			return nil, errors.New("error")
		}

		err = client.ReadWriteWithCheck(mongodb.UsersCollectionID, []byte("key1"), checker)
		require.Nil(mt, err)
	})
}

func TestMongoDBClient_ConcurrentCalls(t *testing.T) {
	t.Parallel()

	client, err := mongodb.CreateMongoDBClient(config.MongoDBConfig{
		URI:    "mongodb://127.0.0.1:27017/?directConnection=true&serverSelectionTimeoutMS=2000",
		DBName: "main",
	})
	require.Nil(t, err)

	checker := func(data interface{}) (interface{}, error) {
		return &core.OTPInfo{}, nil
	}

	numCalls := 600

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
				// _, err := client.IncrementWithTransaction(mongodb.UsersCollectionID, []byte("key"))
				// require.Nil(t, err)
			default:
				assert.Fail(t, "should not hit default")
			}
			wg.Done()
		}(i)
	}
	wg.Wait()
}
