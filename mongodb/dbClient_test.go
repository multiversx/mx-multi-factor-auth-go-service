package mongodb_test

import (
	"errors"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var expectedErr = errors.New("expected error")

func TestNewMongoDBClient(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("nil client, should fail", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(nil, "dbName")
		require.Nil(mt, client)
		require.Equal(mt, mongodb.ErrNilMongoDBClient, err)
	})

	mt.Run("empty db name, should fail", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "")
		require.Nil(mt, client)
		require.Equal(mt, mongodb.ErrEmptyMongoDBName, err)
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateSuccessResponse(),
			mtest.CreateSuccessResponse(),
		)

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

	mt.Run("no documents found, will return ErrKeyNotFound", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: mongo.ErrNoDocuments.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		_, err = client.Get(mongodb.UsersCollectionID, []byte("key1"))
		require.Equal(mt, storage.ErrKeyNotFound.Error(), err.Error())
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

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		val, err := client.IncrementIndex("another coll", []byte("key1"))
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
		require.Equal(mt, uint32(0), val)
	})

	mt.Run("failed to decode entry", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: expectedErr.Error(),
			}),
		)

		val, err := client.IncrementIndex(mongodb.UsersCollectionID, []byte("key1"))
		require.Equal(mt, expectedErr.Error(), err.Error())
		require.Equal(mt, uint32(0), val)
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: bson.D{{Key: "value", Value: 4}}},
		})

		val, err := client.IncrementIndex(mongodb.UsersCollectionID, []byte("key1"))
		require.Nil(mt, err)
		require.Equal(mt, uint32(4), val)
	})
}

func TestMongoDBClient_GetIndex(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		val, err := client.GetIndex("another coll", []byte("key1"))
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
		require.Equal(mt, uint32(0), val)
	})

	mt.Run("failed to decode entry", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: expectedErr.Error(),
			}),
		)

		val, err := client.GetIndex(mongodb.UsersCollectionID, []byte("key1"))
		require.Equal(mt, expectedErr.Error(), err.Error())
		require.Equal(mt, uint32(0), val)
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName")
		require.Nil(mt, err)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key2"},
				{Key: "value", Value: 2},
			}),
		)

		val, err := client.GetIndex(mongodb.UsersCollectionID, []byte("key1"))
		require.Nil(mt, err)
		require.Equal(mt, uint32(2), val)
	})
}
