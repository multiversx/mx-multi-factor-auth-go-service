package core

import (
	"math"

	"github.com/ElrondNetwork/elrond-go-core/core"
)

type bucketIDProvider struct {
	maskHigh        uint32
	maskLow         uint32
	numberOfBuckets uint32
	bytesNeeded     int
}

// NewBucketIDProvider returns a new instance of bucket id provider
func NewBucketIDProvider(numberOfBuckets uint32) (*bucketIDProvider, error) {
	if numberOfBuckets == 0 {
		return nil, ErrInvalidNumberOfBuckets
	}

	provider := &bucketIDProvider{
		numberOfBuckets: numberOfBuckets,
	}
	provider.calculateMasks()
	provider.calculateBytesNeeded()

	return provider, nil
}

// GetIDFromAddress returns the bucket id an address belongs to
func (provider *bucketIDProvider) GetIDFromAddress(address []byte) uint32 {
	if core.IsEmptyAddress(address) {
		return 0
	}
	if provider.bytesNeeded == 0 {
		return 0
	}

	startingIndex := 0
	if len(address) > provider.bytesNeeded {
		startingIndex = len(address) - provider.bytesNeeded
	}

	buffNeeded := address[startingIndex:]
	if core.IsSmartContractOnMetachain(buffNeeded, address) {
		return core.MetachainShardId
	}

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
// ones needed to be taken into consideration for the shard assignment. The result
// of a bitwise AND operation of an address with this mask will result in the
// shard id where a transaction from that address will be dispatched
func (provider *bucketIDProvider) calculateMasks() {
	n := math.Ceil(math.Log2(float64(provider.numberOfBuckets)))
	provider.maskHigh, provider.maskLow = (1<<uint(n))-1, (1<<uint(n-1))-1
}

// calculateBytesNeeded will calculate the next power of 2 starting from the
// number of buckets, then return its power
func (provider *bucketIDProvider) calculateBytesNeeded() {
	n := provider.numberOfBuckets
	n--
	n |= n >> 1
	n |= n >> 2
	n |= n >> 4
	n |= n >> 8
	n |= n >> 16
	n++
	provider.bytesNeeded = int(math.Log2(float64(n)))
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *bucketIDProvider) IsInterfaceNil() bool {
	return provider == nil
}
