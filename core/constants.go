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

const (
	// RedisInstanceConnType specifies a redis connection to a single instance
	RedisInstanceConnType string = "instance"

	// RedisSentinelConnType specifies a redis connection to a setup with sentinel
	RedisSentinelConnType string = "sentinel"
)
