package redis_test

import (
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/multiversx/multi-factor-auth-go-service/redis"
	"github.com/stretchr/testify/require"
)

func TestNewRateLimiter(t *testing.T) {
	t.Parallel()

	t.Run("nil redis client", func(t *testing.T) {
		t.Parallel()

		rl, err := redis.NewRateLimiter(nil)
		require.Nil(t, rl)
		require.Equal(t, redis.ErrNilRedisClient, err)
	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		client, _ := redismock.NewClientMock()

		rl, err := redis.NewRateLimiter(client)
		require.NotNil(t, rl)
		require.False(t, rl.IsInterfaceNil())
		require.Nil(t, err)
	})
}
