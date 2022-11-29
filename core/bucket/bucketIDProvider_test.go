package bucket

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
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
	t.Run("1 bucket  should work", testGetBucketForAddress(1))
	t.Run("4 buckets should work", testGetBucketForAddress(4))
	t.Run("8 buckets should work", testGetBucketForAddress(8))
}

func TestBucketIDProvider_calculateBitsNeeded(t *testing.T) {
	t.Parallel()

	t.Run(" 2 buckets need 1 bit ", testCalculateBitsNeeded(2, 1))
	t.Run(" 3 buckets need 2 bits", testCalculateBitsNeeded(3, 2))
	t.Run(" 4 buckets need 2 bits", testCalculateBitsNeeded(4, 2))
	t.Run(" 5 buckets need 3 bits", testCalculateBitsNeeded(5, 3))
	t.Run(" 6 buckets need 3 bits", testCalculateBitsNeeded(6, 3))
	t.Run(" 7 buckets need 3 bits", testCalculateBitsNeeded(7, 3))
	t.Run(" 8 buckets need 3 bits", testCalculateBitsNeeded(8, 3))
	t.Run("16 buckets need 4 bits", testCalculateBitsNeeded(16, 4))
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

func testCalculateBitsNeeded(numberOfBuckets uint32, expectedBytesNeeded int) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		provider, _ := NewBucketIDProvider(numberOfBuckets)
		assert.False(t, check.IfNil(provider))
		provider.calculateBitsNeeded()
		assert.Equal(t, expectedBytesNeeded, provider.bitsNeeded)
	}
}
