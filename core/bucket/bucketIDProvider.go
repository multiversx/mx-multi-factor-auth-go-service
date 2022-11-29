package bucket

import (
	"math"

	erdCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
)

type bucketIDProvider struct {
	maskHigh        uint32
	maskLow         uint32
	numberOfBuckets uint32
	bitsNeeded      int
}

// NewBucketIDProvider returns a new instance of bucket id provider
func NewBucketIDProvider(numberOfBuckets uint32) (*bucketIDProvider, error) {
	if numberOfBuckets == 0 {
		return nil, core.ErrInvalidNumberOfBuckets
	}

	provider := &bucketIDProvider{
		numberOfBuckets: numberOfBuckets,
	}
	provider.calculateMasks()
	provider.calculateBitsNeeded()
	if provider.bitsNeeded == 0 {
		return nil, core.ErrInvalidNumberOfBuckets
	}

	return provider, nil
}

// GetBucketForAddress returns the bucket id an address belongs to
func (provider *bucketIDProvider) GetBucketForAddress(address []byte) uint32 {
	if erdCore.IsEmptyAddress(address) {
		return 0
	}

	startingIndex := 0
	if len(address) > provider.bitsNeeded {
		startingIndex = len(address) - provider.bitsNeeded
	}

	buffNeeded := address[startingIndex:]
	addr := uint32(0)
	for i := 0; i < len(buffNeeded); i++ {
		addr = addr<<8 + uint32(buffNeeded[i])
	}

	shard := addr & provider.maskHigh
	if shard > provider.numberOfBuckets-1 {
		shard = addr & provider.maskLow
	}

	return shard
}

// calculateMasks will create two numbers whose binary form is composed of as many
// ones needed to be taken into consideration for the bucket assignment. The result
// of a bitwise AND operation of an address with this mask will result in the
// bucket id for the address
func (provider *bucketIDProvider) calculateMasks() {
	n := math.Ceil(math.Log2(float64(provider.numberOfBuckets)))
	provider.maskHigh, provider.maskLow = (1<<uint(n))-1, (1<<uint(n-1))-1
}

// calculateBitsNeeded will calculate the next power of 2 starting from the number of buckets
func (provider *bucketIDProvider) calculateBitsNeeded() {
	n := provider.numberOfBuckets
	if n == 1 {
		provider.bitsNeeded = 1
		return
	}
	provider.bitsNeeded = int(math.Ceil(math.Log2(float64(n))))
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *bucketIDProvider) IsInterfaceNil() bool {
	return provider == nil
}
