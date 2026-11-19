package mongodb_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/storage"
	"github.com/multiversx/mx-multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/mx-multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

var errExpected = errors.New("expected error")

var usersCollID = mongodb.CollectionID(fmt.Sprintf("%s_%d", string(mongodb.UsersCollectionID), 0))

func TestNewMongoDBClient(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("nil client, should fail", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(nil, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, client)
		require.Equal(mt, mongodb.ErrNilMongoDBClient, err)
	})

	mt.Run("empty db name, should fail", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, client)
		require.Equal(mt, mongodb.ErrEmptyMongoDBName, err)
	})

	mt.Run("invalid num of users collections, should fail", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 0, &testscommon.StatusMetricsStub{})
		require.Nil(mt, client)
		require.True(mt, errors.Is(err, core.ErrInvalidValue))
	})

	mt.Run("nil status metrics handler, should fail", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, nil)
		require.Nil(mt, client)
		require.True(mt, errors.Is(err, core.ErrNilMetricsHandler))
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateSuccessResponse(),
			mtest.CreateSuccessResponse(),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
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

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.Put("another coll", []byte("key1"), []byte("data"))
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
	})

	mt.Run("should fail", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: errExpected.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.Put(usersCollID, []byte("key1"), []byte("data"))
		require.Equal(mt, errExpected.Error(), err.Error())
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key2"},
				{Key: "value", Value: []byte("value")},
			}))

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.Put(usersCollID, []byte("key1"), []byte("data"))
		require.Nil(mt, err)
	})
}

func TestMongoDBClient_PutIndexIfNotExists(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.PutIndexIfNotExists("another coll", []byte("key1"), 1)
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
	})

	mt.Run("should fail", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: errExpected.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.PutIndexIfNotExists(usersCollID, []byte("key1"), 1)
		require.Equal(mt, errExpected.Error(), err.Error())
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key2"},
				{Key: "value", Value: []byte("value")},
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.PutIndexIfNotExists(usersCollID, []byte("key1"), 1)
		require.Nil(mt, err)
	})
}

func TestMongoDBClient_Get(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		_, err = client.Get("another coll", []byte("key1"))
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
	})

	mt.Run("find one entry failed", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: errExpected.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		_, err = client.Get(usersCollID, []byte("key1"))
		require.Equal(mt, errExpected.Error(), err.Error())
	})

	mt.Run("no documents found, will return ErrKeyNotFound", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: mongo.ErrNoDocuments.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		_, err = client.Get(usersCollID, []byte("key1"))
		require.Equal(mt, storage.ErrKeyNotFound.Error(), err.Error())
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key2"},
				{Key: "value", Value: []byte("value")},
			}))

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		_, err = client.Get(usersCollID, []byte("key1"))
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
				Message: errExpected.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.Has(usersCollID, []byte("key1"))
		require.Equal(mt, errExpected.Error(), err.Error())
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key2"},
				{Key: "value", Value: []byte("value")},
			}))

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.Has(usersCollID, []byte("key1"))
		require.Nil(mt, err)
	})
}

func TestMongoDBClient_Remove(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
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

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.Remove(usersCollID, []byte("key1"))
		require.Nil(mt, err)
	})
}

func TestMongoDBClient_IncrementIndex(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		val, err := client.IncrementIndex("another coll", []byte("key1"))
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
		require.Equal(mt, uint32(0), val)
	})

	mt.Run("failed to decode entry", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: errExpected.Error(),
			}),
		)

		val, err := client.IncrementIndex(usersCollID, []byte("key1"))
		require.Equal(mt, errExpected.Error(), err.Error())
		require.Equal(mt, uint32(0), val)
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		mt.AddMockResponses(bson.D{
			{Key: "ok", Value: 1},
			{Key: "value", Value: bson.D{{Key: "value", Value: 4}}},
		})

		val, err := client.IncrementIndex(usersCollID, []byte("key1"))
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

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		val, err := client.GetIndex("another coll", []byte("key1"))
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
		require.Equal(mt, uint32(0), val)
	})

	mt.Run("failed to decode entry", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: errExpected.Error(),
			}),
		)

		val, err := client.GetIndex(usersCollID, []byte("key1"))
		require.Equal(mt, errExpected.Error(), err.Error())
		require.Equal(mt, uint32(0), val)
	})

	mt.Run("should work", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key2"},
				{Key: "value", Value: 2},
			}),
		)

		val, err := client.GetIndex(usersCollID, []byte("key1"))
		require.Nil(mt, err)
		require.Equal(mt, uint32(2), val)
	})
}

func TestMongoDBClient_GetWithVersion(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		_, _, err = client.GetWithVersion("another coll", []byte("key1"))
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
	})

	mt.Run("find one failed", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: errExpected.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		_, _, err = client.GetWithVersion(usersCollID, []byte("key1"))
		require.Equal(mt, errExpected.Error(), err.Error())
	})

	mt.Run("no documents found, will return ErrKeyNotFound", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateCursorResponse(0, "foo.bar", mtest.FirstBatch))

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		_, _, err = client.GetWithVersion(usersCollID, []byte("key1"))
		require.Equal(mt, storage.ErrKeyNotFound, err)
	})

	mt.Run("legacy document without version returns version 0", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key1"},
				{Key: "value", Value: []byte("value")},
			}))

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		data, version, err := client.GetWithVersion(usersCollID, []byte("key1"))
		require.Nil(mt, err)
		require.Equal(mt, []byte("value"), data)
		require.Equal(mt, int64(0), version)
	})

	mt.Run("should return the stored version", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
				{Key: "_id", Value: "key1"},
				{Key: "value", Value: []byte("value")},
				{Key: "version", Value: int64(7)},
			}))

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		data, version, err := client.GetWithVersion(usersCollID, []byte("key1"))
		require.Nil(mt, err)
		require.Equal(mt, []byte("value"), data)
		require.Equal(mt, int64(7), version)
	})
}

func TestMongoDBClient_Insert(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.Insert("another coll", []byte("key1"), []byte("data"))
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
	})

	mt.Run("duplicate key should return ErrKeyAlreadyExists", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateWriteErrorsResponse(mtest.WriteError{
			Index:   0,
			Code:    11000,
			Message: "E11000 duplicate key error",
		}))

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.Insert(usersCollID, []byte("key1"), []byte("data"))
		require.Equal(mt, storage.ErrKeyAlreadyExists, err)
	})

	mt.Run("other error should be returned", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: errExpected.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.Insert(usersCollID, []byte("key1"), []byte("data"))
		require.Equal(mt, errExpected.Error(), err.Error())
	})

	mt.Run("should work and set the initial version", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateSuccessResponse())

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.Insert(usersCollID, []byte("key1"), []byte("data"))
		require.Nil(mt, err)

		docs := mt.GetStartedEvent().Command.Lookup("documents").Array()
		doc := docs.Index(0).Value().Document()
		require.Equal(mt, "key1", doc.Lookup("_id").StringValue())
		require.Equal(mt, int64(1), doc.Lookup("version").Int64())
	})
}

func TestMongoDBClient_PutIfVersion(t *testing.T) {
	t.Parallel()

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	mt.Run("collection not found", func(mt *mtest.T) {
		mt.Parallel()

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.PutIfVersion("another coll", []byte("key1"), []byte("data"), 1)
		require.Equal(mt, mongodb.ErrCollectionNotFound, err)
	})

	mt.Run("update error should be returned", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    1,
				Message: errExpected.Error(),
			}),
		)

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.PutIfVersion(usersCollID, []byte("key1"), []byte("data"), 1)
		require.Equal(mt, errExpected.Error(), err.Error())
	})

	mt.Run("no match should return ErrVersionMismatch", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 0}, bson.E{Key: "nModified", Value: 0}))

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.PutIfVersion(usersCollID, []byte("key1"), []byte("data"), 3)
		require.Equal(mt, storage.ErrVersionMismatch, err)
	})

	mt.Run("match should work, filter on version, increment it and never upsert", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1}))

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.PutIfVersion(usersCollID, []byte("key1"), []byte("data"), 3)
		require.Nil(mt, err)

		update := mt.GetStartedEvent().Command.Lookup("updates").Array().Index(0).Value().Document()
		require.Equal(mt, int64(3), update.Lookup("q", "version").Int64())
		require.Equal(mt, int64(1), update.Lookup("u", "$inc", "version").Int64())
		_, setValue := update.Lookup("u", "$set", "value").Binary()
		require.Equal(mt, []byte("data"), setValue)
		upsert, err := update.LookupErr("upsert")
		if err == nil {
			require.False(mt, upsert.Boolean())
		}
	})

	mt.Run("version 0 should match documents without version", func(mt *mtest.T) {
		mt.Parallel()

		mt.AddMockResponses(mtest.CreateSuccessResponse(bson.E{Key: "n", Value: 1}, bson.E{Key: "nModified", Value: 1}))

		client, err := mongodb.NewClient(mt.Client, "dbName", 4, &testscommon.StatusMetricsStub{})
		require.Nil(mt, err)

		err = client.PutIfVersion(usersCollID, []byte("key1"), []byte("data"), 0)
		require.Nil(mt, err)

		update := mt.GetStartedEvent().Command.Lookup("updates").Array().Index(0).Value().Document()
		in := update.Lookup("q", "version", "$in").Array()
		values, err := in.Values()
		require.Nil(mt, err)
		require.Equal(mt, 2, len(values))
		require.Equal(mt, bson.TypeNull, values[0].Type)
		require.Equal(mt, int64(0), values[1].Int64())
	})
}
