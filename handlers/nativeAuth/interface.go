package nativeAuth

// TimestampsCacher defines the behaviour of the block timestamps cacher
type TimestampsCacher interface {
	Get(key []byte) (value interface{}, ok bool)
	Put(key []byte, value interface{}, sizeInBytes int) (evicted bool)
	IsInterfaceNil() bool
}
