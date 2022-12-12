package bucket

import (
	"math"

	erdCore "github.com/ElrondNetwork/elrond-go-core/core"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
)

const bitsPerByte = 8

type bucketIDProvider struct {
	maskHigh        uint32
	maskLow         uint32
	numberOfBuckets uint32
	bytesNeeded     int
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
	provider.calculateBytesNeeded()

	return provider, nil
}

// GetBucketForAddress returns the bucket id an address belongs to
func (provider *bucketIDProvider) GetBucketForAddress(address []byte) uint32 {
	if erdCore.IsEmptyAddress(address) {
		return 0
	}

	startingIndex := 0
	if len(address) > provider.bytesNeeded {
		startingIndex = len(address) - provider.bytesNeeded
	}

	buffNeeded := address[startingIndex:]
	addr := uint32(0)
	for i := 0; i < len(buffNeeded); i++ {
		addr = addr<<8 + uint32(buffNeeded[i])
	}

	bucketIndex := addr & provider.maskHigh
	if bucketIndex > provider.numberOfBuckets-1 {
		bucketIndex = addr & provider.maskLow
	}

	return bucketIndex
}

// calculateMasks will create two numbers whose binary form is composed of as many
// ones needed to be taken into consideration for the bucket assignment. The result
// of a bitwise AND operation of an address with this mask will result in the
// bucket id for the address
func (provider *bucketIDProvider) calculateMasks() {
	n := math.Ceil(math.Log2(float64(provider.numberOfBuckets)))
	provider.maskHigh, provider.maskLow = (1<<uint(n))-1, (1<<uint(n-1))-1
}

// calculateBytesNeeded will calculate the number of bytes needed from an address for index calculation
func (provider *bucketIDProvider) calculateBytesNeeded() {
	n := provider.numberOfBuckets
	if n == 1 {
		provider.bytesNeeded = 1
		return
	}
	provider.bytesNeeded = int(math.Floor(math.Log2(float64(n-1))))/bitsPerByte + 1
}

// IsInterfaceNil returns true if there is no value under the interface
func (provider *bucketIDProvider) IsInterfaceNil() bool {
	return provider == nil
}
