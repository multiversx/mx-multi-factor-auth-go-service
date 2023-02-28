package factory

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	"github.com/multiversx/multi-factor-auth-go-service/config"
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/handlers/storage"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("factory")

const (
	pongValue = "PONG"
)

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

var errInvalidRedisConnType = errors.New("invalid redis connection type")
var errRedisConnectionFailed = errors.New("redis connection failed")

func CreateRedisStorerHandler(cfg config.RedisConfig) (core.Storer, error) {
	switch RedisConnType(cfg.ConnectionType) {
	case RedisInstanceConnType:
		log.Debug("creating redis instance connection")
		return createSimpleClient(cfg)
	case RedisSentinelConnType:
		log.Debug("creating redis sentinel connection")
		return createFailoverClient(cfg)
	case RedisClusterConnType:
		log.Debug("creating redis cluster connection")
		return createClusterClient(cfg)
	default:
		return nil, errInvalidRedisConnType
	}
}

// createSimpleClient will create a redis client for a redis setup with one instance
func createSimpleClient(cfg config.RedisConfig) (core.Storer, error) {
	opt, err := redis.ParseURL(cfg.URL)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(opt)

	ok := isConnected(client)
	if !ok {
		return nil, errRedisConnectionFailed
	}

	rc, err := storage.NewRedisStorerHandler(client)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

// createFailoverClient will create a redis client for a redis setup with sentinel
func createFailoverClient(cfg config.RedisConfig) (core.Storer, error) {
	client := redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:    cfg.MasterName,
		SentinelAddrs: []string{cfg.SentinelURL},
	})

	ok := isConnected(client)
	if !ok {
		return nil, errRedisConnectionFailed
	}

	rc, err := storage.NewRedisStorerHandler(client)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func createClusterClient(cfg config.RedisConfig) (core.Storer, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs: cfg.ClusterAddrs,

		// To route commands by latency or randomly, enable one of the following.
		//RouteByLatency: true,
		//RouteRandomly: true,
	})

	pong, err := client.Ping(context.Background()).Result()
	if err == nil && pong == pongValue {
		return nil, errRedisConnectionFailed
	}

	rc, err := storage.NewRedisStorerHandler(client)
	if err != nil {
		return nil, err
	}

	return rc, nil
}

func isConnected(rc *redis.Client) bool {
	pong, err := rc.Ping(context.Background()).Result()
	return err == nil && pong == pongValue
}
