package mongodb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	"github.com/multiversx/multi-factor-auth-go-service/testscommon"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type testStruct struct {
	Key   string `bson:"_id"`
	Value []byte `bson:"value"`
}

func TestNewMongoDBClient(t *testing.T) {
	t.Parallel()

	t.Run("nil client wrapper, should fail", func(t *testing.T) {
		t.Parallel()

		client, err := mongodb.NewClient(nil, "dbName")
		require.Nil(t, client)
		require.Equal(t, mongodb.ErrNilMongoDBClientWrapper, err)
	})

	t.Run("empty db name, should fail", func(t *testing.T) {
		t.Parallel()

		client, err := mongodb.NewClient(&testscommon.MongoDBClientWrapperStub{}, "")
		require.Nil(t, client)
		require.Equal(t, mongodb.ErrEmptyMongoDBName, err)
	})

	t.Run("failed to connect", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected err")
		mongoDBClientWrapper := &testscommon.MongoDBClientWrapperStub{
			ConnectCalled: func(ctx context.Context) error {
				return expectedErr
			},
		}

		client, err := mongodb.NewClient(mongoDBClientWrapper, "dbName")
		require.Nil(t, client)
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		client, err := mongodb.NewClient(&testscommon.MongoDBClientWrapperStub{}, "dbName")
		require.Nil(t, err)
		require.False(t, client.IsInterfaceNil())
	})
}

func TestMongoDBClient_Put(t *testing.T) {
	t.Parallel()

	t.Run("collection not found", func(t *testing.T) {
		t.Parallel()

		client, err := mongodb.NewClient(&testscommon.MongoDBClientWrapperStub{}, "dbName")
		require.Nil(t, err)

		err = client.Put("another coll", []byte("key1"), []byte("data"))
		require.Equal(t, mongodb.ErrCollectionNotFound, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		mongoDBClientWrapper := &testscommon.MongoDBClientWrapperStub{
			DBCollectionCalled: func(dbName, collName string) mongodb.MongoDBCollection {
				require.Equal(t, string(mongodb.UsersCollectionID), collName)

				return &testscommon.MongoDBCollectionStub{
					UpdateOneCalled: func(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
						wasCalled = true
						return nil, nil
					},
				}
			},
		}

		client, err := mongodb.NewClient(mongoDBClientWrapper, "dbName")
		require.Nil(t, err)

		err = client.Put(mongodb.UsersCollectionID, []byte("key1"), []byte("data"))
		require.Nil(t, err)

		require.True(t, wasCalled)
	})
}

func TestMongoDBClient_Get(t *testing.T) {
	t.Parallel()

	t.Run("collection not found", func(t *testing.T) {
		t.Parallel()

		client, err := mongodb.NewClient(&testscommon.MongoDBClientWrapperStub{}, "dbName")
		require.Nil(t, err)

		_, err = client.Get("another coll", []byte("key1"))
		require.Equal(t, mongodb.ErrCollectionNotFound, err)
	})

	t.Run("find one entry failed", func(t *testing.T) {
		t.Parallel()

		expectedErr := errors.New("expected err")
		mongoDBClientWrapper := &testscommon.MongoDBClientWrapperStub{
			DBCollectionCalled: func(dbName, collName string) mongodb.MongoDBCollection {
				require.Equal(t, string(mongodb.UsersCollectionID), collName)

				return &testscommon.MongoDBCollectionStub{
					FindOneCalled: func(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
						return mongo.NewSingleResultFromDocument(&testStruct{}, expectedErr, bson.DefaultRegistry)
					},
				}
			},
		}

		client, err := mongodb.NewClient(mongoDBClientWrapper, "dbName")
		require.Nil(t, err)

		_, err = client.Get(mongodb.UsersCollectionID, []byte("key1"))
		require.Equal(t, expectedErr, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mongoDBClientWrapper := &testscommon.MongoDBClientWrapperStub{
			DBCollectionCalled: func(dbName, collName string) mongodb.MongoDBCollection {
				require.Equal(t, string(mongodb.UsersCollectionID), collName)

				return &testscommon.MongoDBCollectionStub{
					FindOneCalled: func(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
						return mongo.NewSingleResultFromDocument(&testStruct{}, nil, bson.DefaultRegistry)
					},
				}
			},
		}

		client, err := mongodb.NewClient(mongoDBClientWrapper, "dbName")
		require.Nil(t, err)

		_, err = client.Get(mongodb.UsersCollectionID, []byte("key1"))
		require.Nil(t, err)
	})
}

func TestMongoDBClient_Has(t *testing.T) {
	t.Parallel()

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		mongoDBClientWrapper := &testscommon.MongoDBClientWrapperStub{
			DBCollectionCalled: func(dbName, collName string) mongodb.MongoDBCollection {
				require.Equal(t, string(mongodb.UsersCollectionID), collName)

				return &testscommon.MongoDBCollectionStub{
					FindOneCalled: func(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
						return mongo.NewSingleResultFromDocument(&testStruct{}, nil, bson.DefaultRegistry)
					},
				}
			},
		}

		client, err := mongodb.NewClient(mongoDBClientWrapper, "dbName")
		require.Nil(t, err)

		err = client.Has(mongodb.UsersCollectionID, []byte("key1"))
		require.Nil(t, err)
	})
}

func TestMongoDBClient_Remove(t *testing.T) {
	t.Parallel()

	t.Run("collection not found", func(t *testing.T) {
		t.Parallel()

		client, err := mongodb.NewClient(&testscommon.MongoDBClientWrapperStub{}, "dbName")
		require.Nil(t, err)

		err = client.Remove("another coll", []byte("key1"))
		require.Equal(t, mongodb.ErrCollectionNotFound, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		wasCalled := false
		mongoDBClientWrapper := &testscommon.MongoDBClientWrapperStub{
			DBCollectionCalled: func(dbName, collName string) mongodb.MongoDBCollection {
				require.Equal(t, string(mongodb.UsersCollectionID), collName)

				return &testscommon.MongoDBCollectionStub{
					DeleteOneCalled: func(ctx context.Context, filter interface{}, opts ...*options.DeleteOptions) (*mongo.DeleteResult, error) {
						wasCalled = true
						return nil, nil
					},
				}
			},
		}

		client, err := mongodb.NewClient(mongoDBClientWrapper, "dbName")
		require.Nil(t, err)

		err = client.Remove(mongodb.UsersCollectionID, []byte("key1"))
		require.Nil(t, err)

		require.True(t, wasCalled)
	})
}

// func TestMongoDBClient_IncrementWithTransaction(t *testing.T) {
// 	t.Parallel()

// 	updateWasCalled := false
// 	findWasCalled := false
// 	sessionWasCalled := false

// 	mongoDBClientWrapper := &testscommon.MongoDBClientWrapperStub{
// 		DBCollectionCalled: func(dbName, collName string) mongodb.MongoDBCollection {
// 			require.Equal(t, string(mongodb.UsersCollectionID), collName)

// 			return &testscommon.MongoDBCollectionStub{
// 				FindOneCalled: func(ctx context.Context, filter interface{}, opts ...*options.FindOneOptions) *mongo.SingleResult {
// 					findWasCalled = true
// 					return mongo.NewSingleResultFromDocument(&testStruct{Key: "key", Value: []byte{0, 0, 0, 1}}, nil, bson.DefaultRegistry)
// 				},
// 				UpdateOneCalled: func(ctx context.Context, filter, update interface{}, opts ...*options.UpdateOptions) (*mongo.UpdateResult, error) {
// 					updateWasCalled = true
// 					return nil, nil
// 				},
// 			}
// 		},
// 		StartSessionCalled: func() (mongo.Session, error) {
// 			sessionWasCalled = true
// 			return &testscommon.MongoDBSessionStub{
// 				WithTransactionCalled: func(ctx context.Context, fn func(sessCtx mongo.SessionContext) (interface{}, error), opts ...*options.TransactionOptions) (interface{}, error) {
// 					fn(mongo.NewSessionContext(context.TODO(), mongo.SessionFromContext(context.TODO())))
// 					return uint32(5), nil
// 				},
// 			}, nil
// 		},
// 	}

// 	client, err := mongodb.NewClient(mongoDBClientWrapper, "dbName")
// 	require.Nil(t, err)

// 	_, err = client.IncrementWithTransaction(mongodb.UsersCollectionID, []byte("key1"))
// 	require.Nil(t, err)

// 	require.True(t, findWasCalled)
// 	require.True(t, updateWasCalled)
// 	require.True(t, sessionWasCalled)
// }

func TestMongoDBClient_Close(t *testing.T) {
	t.Parallel()

	wasCalled := false
	mongoDBClientWrapper := &testscommon.MongoDBClientWrapperStub{
		ConnectCalled: func(ctx context.Context) error {
			wasCalled = true
			return nil
		},
	}

	client, err := mongodb.NewClient(mongoDBClientWrapper, "dbName")
	require.Nil(t, err)

	require.Nil(t, client.Close())
	require.True(t, wasCalled)
}
