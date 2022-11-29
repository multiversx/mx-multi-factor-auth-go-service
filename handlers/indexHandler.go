package handlers

import (
	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
)

const indexMultiplier = 2

// ArgIndexHandler is the DTO used to create a new instance of index handler
type ArgIndexHandler struct {
	BucketIndexHolder core.BucketIndexHolder
	BucketIDProvider  core.BucketIDProvider
}

type indexHandler struct {
	bucketIndexHolder core.BucketIndexHolder
	bucketIDProvider  core.BucketIDProvider
}

// NewIndexHandler returns a new instance of index handler
func NewIndexHandler(args ArgIndexHandler) (*indexHandler, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &indexHandler{
		bucketIndexHolder: args.BucketIndexHolder,
		bucketIDProvider:  args.BucketIDProvider,
	}, nil
}

func checkArgs(args ArgIndexHandler) error {
	if check.IfNil(args.BucketIndexHolder) {
		return ErrNilBucketIndexHolder
	}
	if check.IfNil(args.BucketIDProvider) {
		return ErrNilBucketIDProvider
	}

	return nil
}

// AllocateIndex returns a new index that was not used before
func (ih *indexHandler) AllocateIndex(address []byte) (uint32, error) {
	baseIndex, err := ih.bucketIndexHolder.UpdateIndexReturningNext(address)
	if err != nil {
		return 0, err
	}

	bucketID := ih.bucketIDProvider.GetBucketForAddress(address)

	return ih.getNextFinalIndex(baseIndex, bucketID), nil
}

func (ih *indexHandler) getNextFinalIndex(newIndex, bucketID uint32) uint32 {
	numBuckets := ih.bucketIndexHolder.NumberOfBuckets()
	return indexMultiplier * (newIndex*numBuckets + bucketID)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ih *indexHandler) IsInterfaceNil() bool {
	return ih == nil
}
