package mongo

import (
	"fmt"
	"sync"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/mongodb"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("handlers/storage/mongo")

type mongodbStorerHandler struct {
	client     mongodb.MongoDBClient
	collection mongodb.CollectionID
	sessions   map[string]mongodb.Session
	sessionCtx map[string]mongodb.SessionContext
	mutSess    sync.RWMutex
}

// NewMongoDBStorerHandler will create a new storer handler instance
func NewMongoDBStorerHandler(client mongodb.MongoDBClient, collection mongodb.CollectionID) (*mongodbStorerHandler, error) {
	if client == nil {
		return nil, core.ErrNilMongoDBClient
	}

	return &mongodbStorerHandler{
		client:     client,
		collection: collection,
		sessions:   make(map[string]mongodb.Session),
		sessionCtx: make(map[string]mongodb.SessionContext),
	}, nil
}

// Put will set key value pair
func (msh *mongodbStorerHandler) Put(key []byte, data []byte) error {
	// return msh.client.Put(msh.collection, key, data)
	log.Debug("StorerM: put", "key", key, "data", data)

	msh.mutSess.Lock()
	defer msh.mutSess.Unlock()

	session, ok := msh.sessions[string(key)]
	if !ok {
		log.Trace("%w: could not find session for key %s", core.ErrInvalidValue, string(key))
		return fmt.Errorf("%w: could not find session for key %s", core.ErrInvalidValue, string(key))
	}
	if session == nil {
		log.Trace("nil session for key %s", string(key))
		return fmt.Errorf("nil session for key %s", string(key))
	}

	sessionCtx, ok := msh.sessionCtx[string(key)]
	if !ok {
		log.Trace("%w: could not find session context for key %s", core.ErrInvalidValue, string(key))
		return fmt.Errorf("%w: could not find session context for key %s", core.ErrInvalidValue, string(key))
	}
	if sessionCtx == nil {
		log.Trace("nil session context for key %s", string(key))
		return fmt.Errorf("nil session context for key %s", string(key))
	}

	err := msh.client.WriteWithTx(msh.collection, key, data, session, sessionCtx)
	if err != nil {
		log.Trace("StorerM: put", "key", key, "err", err.Error())
		return err
	}

	return nil
}

// Get will return the value for the provided key
func (msh *mongodbStorerHandler) Get(key []byte) ([]byte, error) {
	//return msh.client.Get(msh.collection, key)
	msh.mutSess.Lock()
	defer msh.mutSess.Unlock()

	data, session, sessionCtx, err := msh.client.ReadWithTx(msh.collection, key)
	msh.sessions[string(key)] = session
	msh.sessionCtx[string(key)] = sessionCtx
	if err != nil {
		log.Debug("StorerM: get", "key", key, "err", err.Error())
		return nil, err
	}

	log.Debug("StorerM: get", "key", key, "data", data)

	return data, nil
}

// Has will return true if the provided key exists in the database collection
func (msh *mongodbStorerHandler) Has(key []byte) error {
	return msh.client.Has(msh.collection, key)
}

// SearchFirst will return the provided key
func (msh *mongodbStorerHandler) SearchFirst(key []byte) ([]byte, error) {
	return msh.Get(key)
}

// Remove will remove the provided key from the database collection
func (msh *mongodbStorerHandler) Remove(key []byte) error {
	return msh.client.Remove(msh.collection, key)
}

// ClearCache is not implemented
func (msh *mongodbStorerHandler) ClearCache() {
	log.Warn("ClearCache: NOT implemented")
}

// Close will close the mongodb client
func (msh *mongodbStorerHandler) Close() error {
	return msh.client.Close()
}

// IsInterfaceNil returns true if there is no value under the interface
func (msh *mongodbStorerHandler) IsInterfaceNil() bool {
	return msh == nil
}
