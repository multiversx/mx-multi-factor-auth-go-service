package core

// WebServerOffString represents the constant used to switch off the web server
const WebServerOffString = "off"

// DBType represents the type of DB
type DBType string

// LevelDB is the local levelDB
const LevelDB DBType = "levelDB"

// MongoDB is the mongo db identifier
const MongoDB DBType = "mongoDB"

const (
	getAccountEndpointFormat      = "address/%s"
	getGuardianDataEndpointFormat = "address/%s/guardian-data"
)

// RedisConnType defines the redis connection type
type RedisConnType string

const (
	// RedisInstanceConnType specifies a redis connection to a single instance
	RedisInstanceConnType RedisConnType = "instance"

	// RedisSentinelConnType specifies a redis connection to a setup with sentinel
	RedisSentinelConnType RedisConnType = "sentinel"
)

// NoExpiryValue is the returned value for a persistent key expiry time
const NoExpiryValue = -1
