package mongodb

// MongoDBClient defines what a mongodb client should do
type MongoDBClient interface {
	Put(coll CollectionID, key []byte, data []byte) error
	Get(coll CollectionID, key []byte) ([]byte, error)
	GetWithVersion(coll CollectionID, key []byte) ([]byte, int64, error)
	Insert(coll CollectionID, key []byte, data []byte) error
	PutIfVersion(coll CollectionID, key []byte, data []byte, expectedVersion int64) error
	Has(coll CollectionID, key []byte) error
	Remove(coll CollectionID, key []byte) error
	GetIndex(collID CollectionID, key []byte) (uint32, error)
	PutIndexIfNotExists(collID CollectionID, key []byte, index uint32) error
	IncrementIndex(collID CollectionID, key []byte) (uint32, error)
	GetAllCollectionsIDs() []CollectionID
	Close() error
	IsInterfaceNil() bool
}
