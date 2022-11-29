package bucket

import (
	"fmt"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	logger "github.com/ElrondNetwork/elrond-go-logger"
	erdCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
)

var log = logger.GetOrCreate("bucket")

// ArgBucketIndexHolder is the DTO used to create a new instance of bucket index holder
type ArgBucketIndexHolder struct {
	BucketIDProvider core.BucketIDProvider
	BucketHandlers   map[uint32]core.BucketIndexHandler
}

type bucketIndexHolder struct {
	BucketIDProvider core.BucketIDProvider
	BucketHandlers   map[uint32]core.BucketIndexHandler
}

// NewBucketIndexHolder returns a new instance of bucket index holder
func NewBucketIndexHolder(args ArgBucketIndexHolder) (*bucketIndexHolder, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &bucketIndexHolder{
		BucketIDProvider: args.BucketIDProvider,
		BucketHandlers:   args.BucketHandlers,
	}, nil
}

func checkArgs(args ArgBucketIndexHolder) error {
	if check.IfNil(args.BucketIDProvider) {
		return core.ErrNilBucketIDProvider
	}
	if len(args.BucketHandlers) == 0 {
		return core.ErrInvalidBucketHandlers
	}

	for id, bucketHandler := range args.BucketHandlers {
		if check.IfNil(bucketHandler) {
			return fmt.Errorf("%w for id %d", core.ErrNilBucketHandler, id)
		}
	}

	return nil
}

// Put adds data to the bucket where the key should be
func (holder *bucketIndexHolder) Put(key, data []byte) error {
	bucket, err := holder.getBucketForAddress(key)
	if err != nil {
		return err
	}

	return bucket.Put(key, data)
}

// Get returns the value for the key from the bucket where the key should be
func (holder *bucketIndexHolder) Get(key []byte) ([]byte, error) {
	bucket, err := holder.getBucketForAddress(key)
	if err != nil {
		return make([]byte, 0), err
	}

	return bucket.Get(key)
}

// Has returns true if the key exists in the bucket where the key should be
func (holder *bucketIndexHolder) Has(key []byte) error {
	bucket, err := holder.getBucketForAddress(key)
	if err != nil {
		return err
	}

	return bucket.Has(key)
}

// Close closes the managed buckets
func (holder *bucketIndexHolder) Close() error {
	var lastError error
	for idx, bucket := range holder.BucketHandlers {
		errClose := bucket.Close()
		if errClose != nil {
			lastError = errClose
			log.Error("could not close bucket, error: %w, index: %idx", lastError, idx)
		}
	}

	return lastError
}

// UpdateIndexReturningNext updates the index for the provided address and returns the new value
func (holder *bucketIndexHolder) UpdateIndexReturningNext(address []byte) (uint32, error) {
	bucket, err := holder.getBucketForAddress(address)
	if err != nil {
		return 0, err
	}

	return bucket.UpdateIndexReturningNext()
}

// NumberOfBuckets returns the total number of buckets
func (holder *bucketIndexHolder) NumberOfBuckets() uint32 {
	return uint32(len(holder.BucketHandlers))
}

func (holder *bucketIndexHolder) getBucketForAddress(address []byte) (core.BucketIndexHandler, error) {
	bucketID := holder.BucketIDProvider.GetBucketForAddress(address)
	bucket, found := holder.BucketHandlers[bucketID]
	if !found {
		return nil, fmt.Errorf("%w for address %s", core.ErrInvalidBucketID, erdCore.AddressPublicKeyConverter.Encode(address))
	}

	return bucket, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (holder *bucketIndexHolder) IsInterfaceNil() bool {
	return holder == nil
}
