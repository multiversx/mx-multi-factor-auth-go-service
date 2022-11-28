package core

import (
	"testing"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestNewBucketIDProvider(t *testing.T) {
	t.Parallel()

	t.Run("invalid number of buckets should error", func(t *testing.T) {
		t.Parallel()

		provider, err := NewBucketIDProvider(0)
		assert.Equal(t, ErrInvalidNumberOfBuckets, err)
		assert.True(t, check.IfNil(provider))
	})
	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		provider, err := NewBucketIDProvider(5)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(provider))
	})
}

func TestBucketIDProvider_GetIDFromAddress(t *testing.T) {
	t.Parallel()

	t.Run("empty address should return 0", func(t *testing.T) {
		t.Parallel()

		provider, err := NewBucketIDProvider(5)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(provider))
		assert.Zero(t, provider.GetIDFromAddress([]byte("")))
	})
	t.Run("1 bucket  should work", testGetIDFromAddress(1))
	t.Run("4 buckets should work", testGetIDFromAddress(4))
	t.Run("8 buckets should work", testGetIDFromAddress(8))
}

func TestBucketIDProvider_calculateBytesNeeded(t *testing.T) {
	t.Parallel()

	t.Run(" 1 bucket needs 1 byte ", testCalculateBytesNeeded(1, 0))
	t.Run(" 2 buckets need 1 byte ", testCalculateBytesNeeded(2, 1))
	t.Run(" 3 buckets need 2 bytes", testCalculateBytesNeeded(3, 2))
	t.Run(" 4 buckets need 2 bytes", testCalculateBytesNeeded(4, 2))
	t.Run(" 5 buckets need 3 bytes", testCalculateBytesNeeded(5, 3))
	t.Run(" 6 buckets need 3 bytes", testCalculateBytesNeeded(6, 3))
	t.Run(" 7 buckets need 3 bytes", testCalculateBytesNeeded(7, 3))
	t.Run(" 8 buckets need 3 bytes", testCalculateBytesNeeded(8, 3))
	t.Run("16 buckets need 4 bytes", testCalculateBytesNeeded(16, 4))
}

func testGetIDFromAddress(numBuckets uint32) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		provider, err := NewBucketIDProvider(numBuckets)
		assert.Nil(t, err)
		assert.False(t, check.IfNil(provider))
		numCalls := 100
		for i := 0; i < numCalls; i++ {
			assert.Equal(t, uint32(i)%numBuckets, provider.GetIDFromAddress([]byte{byte(i)}))
		}
	}
}

func testCalculateBytesNeeded(numberOfBuckets uint32, expectedBytesNeeded int) func(t *testing.T) {
	return func(t *testing.T) {
		t.Parallel()

		provider, _ := NewBucketIDProvider(numberOfBuckets)
		assert.False(t, check.IfNil(provider))
		provider.calculateBytesNeeded()
		assert.Equal(t, expectedBytesNeeded, provider.bytesNeeded)
	}
}
