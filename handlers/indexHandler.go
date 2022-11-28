package handlers

import (
	"encoding/binary"
	"fmt"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	erdCore "github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
)

const (
	lastIndexKey = "lastAllocatedIndex"
	uint32Bytes  = 4
)

// ArgIndexHandler is the DTO used to create a new instance of index handler
type ArgIndexHandler struct {
	BucketIDProvider core.BucketIDProvider
	IndexBuckets     map[uint32]core.Storer
}

type indexHandler struct {
	bucketIDProvider core.BucketIDProvider
	indexBuckets     map[uint32]core.Storer
	bucketsMutexes   map[uint32]*sync.Mutex
}

// NewIndexHandler returns a new instance of index handler
func NewIndexHandler(args ArgIndexHandler) (*indexHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	ih := &indexHandler{
		bucketIDProvider: args.BucketIDProvider,
		indexBuckets:     args.IndexBuckets,
	}
	ih.initEmptyBuckets()
	ih.initMutexes()

	return ih, err
}

func checkArgs(args ArgIndexHandler) error {
	if check.IfNil(args.BucketIDProvider) {
		return ErrNilBucketIDProvider
	}
	if len(args.IndexBuckets) == 0 {
		return InvalidNumberOfBuckets
	}

	nilDBsCounter := 0
	for _, db := range args.IndexBuckets {
		if check.IfNil(db) {
			nilDBsCounter++
		}
	}
	if nilDBsCounter != 0 {
		return fmt.Errorf("%w for %d databases", ErrNilDB, nilDBsCounter)
	}

	return nil
}

// AllocateIndex returns a new index that was not used before
func (ih *indexHandler) AllocateIndex(address []byte) (uint32, error) {
	bucketID := ih.bucketIDProvider.GetIDFromAddress(address)
	mut, found := ih.bucketsMutexes[bucketID]
	if !found {
		return 0, fmt.Errorf("%w for address %s", ErrInvalidBucketID, erdCore.AddressPublicKeyConverter.Encode(address))
	}

	mut.Lock()
	defer mut.Unlock()

	bucket, found := ih.indexBuckets[bucketID]
	if !found {
		return 0, ErrInvalidBucketID
	}

	lastBaseIndex, err := ih.getIndex(bucket)
	if err != nil {
		return 0, err
	}
	lastBaseIndex++

	err = ih.saveNewIndex(lastBaseIndex, bucket)
	if err != nil {
		return 0, err
	}

	return ih.getNextFinalIndex(lastBaseIndex, bucketID), nil
}

func (ih *indexHandler) getNextFinalIndex(newIndex uint32, bucketID uint32) uint32 {
	numBuckets := uint32(len(ih.indexBuckets))
	return newIndex*numBuckets + bucketID
}

func (ih *indexHandler) initEmptyBuckets() {
	for _, bucket := range ih.indexBuckets {
		err := bucket.Has([]byte(lastIndexKey))
		if err != nil {
			err = ih.saveNewIndex(0, bucket)
		}
	}
}

func (ih *indexHandler) initMutexes() {
	numberOfBuckets := len(ih.indexBuckets)
	ih.bucketsMutexes = make(map[uint32]*sync.Mutex, numberOfBuckets)
	for i := 0; i < numberOfBuckets; i++ {
		ih.bucketsMutexes[uint32(i)] = &sync.Mutex{}
	}
}

func (ih *indexHandler) getIndex(bucket core.Storer) (uint32, error) {
	lastIndexBytes, err := bucket.Get([]byte(lastIndexKey))
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(lastIndexBytes), nil
}

func (ih *indexHandler) saveNewIndex(newIndex uint32, bucket core.Storer) error {
	latestIndexBytes := make([]byte, uint32Bytes)
	binary.BigEndian.PutUint32(latestIndexBytes, newIndex)
	return bucket.Put([]byte(lastIndexKey), latestIndexBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ih *indexHandler) IsInterfaceNil() bool {
	return ih == nil
}
