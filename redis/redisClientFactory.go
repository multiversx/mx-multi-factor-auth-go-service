package redis

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/multiversx/multi-factor-auth-go-service/config"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("redis")

const (
	pongValue = "PONG"
)

var errInvalidRedisConnType = errors.New("invalid redis connection type")
var errRedisConnectionFailed = errors.New("redis connection failed")

// RedisConnType defines redis connection type
type RedisConnType string

const (
	// RedisInstanceConnType specifies a redis connection to a single instance
	RedisInstanceConnType RedisConnType = "instance"

	// RedisSentinelConnType specifies a redis connection to a setup with sentinel
	RedisSentinelConnType RedisConnType = "sentinel"

	// RedisClusterConnType specifies a redis connection to a setup with a cluster of nodes
	RedisClusterConnType RedisConnType = "cluster"
)

// CreateRedisClient will create a new redis client wrapper
func CreateRedisClient(cfg config.RedisConfig) (RedisClientWrapper, error) {
	switch RedisConnType(cfg.ConnectionType) {
	case RedisInstanceConnType:
		return createSimpleClient(cfg)
	case RedisSentinelConnType:
		return createFailoverClient(cfg)
	case RedisClusterConnType:
		return createClusterClient(cfg)
	default:
		return nil, errInvalidRedisConnType
	}
}

// createSimpleClient will create a redis client for a redis setup with one instance
func createSimpleClient(cfg config.RedisConfig) (RedisClientWrapper, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)

	log.Debug("creating redis instance connection")

	return createRedisClientWrapper(client)
}

// createFailoverClient will create a redis client for a redis setup with sentinel
func createFailoverClient(cfg config.RedisConfig) (RedisClientWrapper, error) {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    cfg.MasterName,
		SentinelAddrs: []string{cfg.SentinelURL},
	})

	log.Debug("creating redis sentinel connection")

	return createRedisClientWrapper(client)
}

// createClusterClient will create a redis client for a redis setup with cluster
func createClusterClient(cfg config.RedisConfig) (RedisClientWrapper, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: cfg.ClusterAddrs,

		// To route commands by latency or randomly, enable one of the following.
		//RouteByLatency: true,
		//RouteRandomly: true,
	})

	log.Debug("creating redis cluster connection")

	return createRedisClientWrapper(client)
}

func createRedisClientWrapper(client RedisClient) (RedisClientWrapper, error) {
	ok := isConnected(client)
	if !ok {
		return nil, errRedisConnectionFailed
	}

	rc, err := NewRedisClientWrapper(client)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func isConnected(rc RedisClient) bool {
	pong, err := rc.Ping(context.Background()).Result()
	return err == nil && pong == pongValue
}
