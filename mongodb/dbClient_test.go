package mongodb_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var expectedErr = errors.New("expected error")

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
			}))

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		err = client.Put(mongodb.UsersCollectionID, []byte("key1"), []byte("data"))
		require.Nil(mt, err)
	})
}

func TestMongoDBClient_PutIndexIfNotExists(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		err = client.PutIndexIfNotExists("another coll", []byte("key1"), 1)
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
	})

	mt.Run("should fail", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: expectedErr.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		err = client.PutIndexIfNotExists(mongodb.UsersCollectionID, []byte("key1"), 1)
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

		err = client.PutIndexIfNotExists(mongodb.UsersCollectionID, []byte("key1"), 1)
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
			}))

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
			}))

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

func TestMongoDBClient_IncrementIndex(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		mt.AddMockResponses(bson.D{
			{"ok", 1},
			{"value", bson.D{{Key: "value", Value: 4}}},
		})

		val, err := client.IncrementIndex(mongodb.UsersCollectionID, []byte("key1"))
		require.Nil(mt, err)
		require.Equal(mt, uint32(4), val)
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
				{Key: "value", Value: []byte("data1")},
			}),
			mtest.CreateSuccessResponse(),
			mtest.CreateSuccessResponse(),
		)

		checker := func(data interface{}) (interface{}, error) {
			if bytes.Equal(data.([]byte), []byte("data1")) {
				return []byte("data2"), nil
			}
			return nil, errors.New("error")
		}

		err = client.ReadWriteWithCheck(mongodb.UsersCollectionID, []byte("key1"), checker)
		require.Nil(mt, err)
	})
}
