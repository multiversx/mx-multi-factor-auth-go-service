package bucket

import (
	"testing"

	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/stretchr/testify/assert"
)

func TestNewBucketIDProvider(t *testing.T) {
	t.Parallel()

	t.Run("invalid number of buckets should error", func(t *testing.T) {
		t.Parallel()

		provider, err := NewBucketIDProvider(0)
		assert.Equal(t, core.ErrInvalidNumberOfBuckets, err)
		assert.True(t, check.IfNil(provider))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		provider, err := NewBucketIDProvider(5)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(provider))
	})
}

func TestBucketIDProvider_GetBucketForAddress(t *testing.T) {
	t.Parallel()

	t.Run("empty address should return 0", func(t *testing.T) {
		t.Parallel()

		provider, err := NewBucketIDProvider(5)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(provider))
		assert.Zero(t, provider.GetBucketForAddress([]byte("")))
	})
	t.Run("real address should work", func(t *testing.T) {
		t.Parallel()

		addr, err := data.NewAddressFromBech32String("erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th")
		assert.Nil(t, err)

		provider, err := NewBucketIDProvider(4)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(provider))
		bucketIndex := provider.GetBucketForAddress(addr.AddressBytes())
		assert.Equal(t, uint32(1), bucketIndex)
	})
	t.Run("1 bucket  should work", testGetBucketForAddress(1))
	t.Run("4 buckets should work", testGetBucketForAddress(4))
	t.Run("8 buckets should work", testGetBucketForAddress(8))
}

func TestBucketIDProvider_calculateBytesNeeded(t *testing.T) {
	t.Parallel()

	t.Run("         1 bucket needs 1 byte ", testCalculateBytesNeeded(1, 1))
	t.Run("        16 buckets need 1 byte ", testCalculateBytesNeeded(16, 1))
	t.Run("       255 buckets need 1 byte ", testCalculateBytesNeeded(255, 1))
	t.Run("       256 buckets need 1 byte ", testCalculateBytesNeeded(256, 1))
	t.Run("     65535 buckets need 2 bytes", testCalculateBytesNeeded(65535, 2))
	t.Run("     65536 buckets need 2 bytes", testCalculateBytesNeeded(65536, 2))
	t.Run("  16777215 buckets need 3 bytes", testCalculateBytesNeeded(16777215, 3))
	t.Run("  16777216 buckets need 3 bytes", testCalculateBytesNeeded(16777216, 3))
	t.Run("4294967295 buckets need 4 bytes", testCalculateBytesNeeded(4294967295, 4))
}

func testGetBucketForAddress(numBuckets uint32) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		provider, err := NewBucketIDProvider(numBuckets)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(provider))
		numCalls := 100
		for i := 0; i < numCalls; i++ {
			assert.Equal(t, uint32(i)%numBuckets, provider.GetBucketForAddress([]byte{byte(i)}))
		}
	}
}

func testCalculateBytesNeeded(numberOfBuckets uint32, expectedBytesNeeded int) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		provider, _ := NewBucketIDProvider(numberOfBuckets)
		assert.False(t, check.IfNil(provider))
		assert.Equal(t, expectedBytesNeeded, provider.bytesNeeded)
	}
}
