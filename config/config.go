package config

import (
	"github.com/multiversx/mx-chain-storage-go/common"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core"
)

// Configs is a holder for the relayer configuration parameters
type Configs struct {
	GeneralConfig   Config
	ExternalConfig  ExternalConfig
	ApiRoutesConfig ApiRoutesConfig
	FlagsConfig     ContextFlagsConfig
}

// Config general configuration struct
type Config struct {
	Guardian         GuardianConfig
	General          GeneralConfig
	Logs             LogsConfig
	Antiflood        AntifloodConfig
	ServiceResolver  ServiceResolverConfig
	ShardedStorage   ShardedStorageConfig
	TwoFactor        TwoFactorConfig
	NativeAuthServer NativeAuthServerConfig
	PubKey           PubkeyConfig
}

// ExternalConfig defines the configuration for external components
type ExternalConfig struct {
	Api     ApiConfig
	MongoDB MongoDBConfig
	Redis   RedisConfig
	Gin     GinConfig
}

// ShardedStorageConfig is the configuration for the sharded storage
type ShardedStorageConfig struct {
	NumberOfBuckets uint32
	Users           StorageConfig
}

// StorageConfig will map the storage unit configuration
type StorageConfig struct {
	Cache common.CacheConfig
	DB    common.DBConfig
}

// ContextFlagsConfig the configuration for flags
type ContextFlagsConfig struct {
	WorkingDir                string
	LogLevel                  string
	ConfigurationFile         string
	ConfigurationApiFile      string
	ConfigurationExternalFile string
	RestApiInterface          string
	DisableAnsiColor          bool
	SaveLogFile               bool
	EnableLogName             bool
	EnablePprof               bool
	StartSwaggerUI            bool
}

// WebServerAntifloodConfig will hold the anti-flooding parameters for the web server
type WebServerAntifloodConfig struct {
	SimultaneousRequests         uint32
	SameSourceRequests           uint32
	SameSourceResetIntervalInSec uint32
}

// AntifloodConfig will hold all p2p antiflood parameters
type AntifloodConfig struct {
	Enabled   bool
	WebServer WebServerAntifloodConfig
}

// ApiRoutesConfig holds the configuration related to Rest API routes
type ApiRoutesConfig struct {
	RestApiInterface string
	Logging          ApiLoggingConfig
	APIPackages      map[string]APIPackageConfig
}

// ApiLoggingConfig holds the configuration related to API requests logging
type ApiLoggingConfig struct {
	LoggingEnabled          bool
	ThresholdInMicroSeconds int
}

// APIPackageConfig holds the configuration for the routes of each package
type APIPackageConfig struct {
	Routes []RouteConfig
}

// RouteConfig holds the configuration for a single route
type RouteConfig struct {
	Name             string
	Open             bool
	Auth             bool
	MaxContentLength uint64
}

// GuardianConfig holds the configuration for the guardian
type GuardianConfig struct {
	MnemonicFile         string
	RequestTimeInSeconds int
}

// GeneralConfig holds the general configuration for the service
type GeneralConfig struct {
	Marshalizer string
	DBType      core.DBType
}

// ApiConfig will hold settings related to the Api
type ApiConfig struct {
	NetworkAddress string
}

// LogsConfig will hold settings related to the logging sub-system
type LogsConfig struct {
	LogFileLifeSpanInSec int
}

// ServiceResolverConfig will hold settings related to the service resolver
type ServiceResolverConfig struct {
	RequestTimeInSeconds             uint64
	SkipTxUserSigVerify              bool
	MaxTransactionsAllowedForSigning int
	DelayBetweenOTPWritesInSec       uint64
}

// TwoFactorConfig will hold settings related to the two factor totp
type TwoFactorConfig struct {
	Issuer                           string
	Digits                           int
	BackoffTimeInSeconds             uint64
	MaxFailures                      int64
	SecurityModeMaxFailures          int64
	SecurityModeBackoffTimeInSeconds uint64
}

// MongoDBConfig maps the mongodb configuration
type MongoDBConfig struct {
	URI                   string
	DBName                string
	ConnectTimeoutInSec   uint32
	OperationTimeoutInSec uint32
	NumUsersCollections   uint32
}

// NativeAuthServerConfig will hold settings related to the native auth server
type NativeAuthServerConfig struct {
	Cache common.CacheConfig
}

// RedisConfig maps the redis configuration
type RedisConfig struct {
	URL                   string
	Channel               string
	MasterName            string
	SentinelUrl           string
	ConnectionType        string
	OperationTimeoutInSec uint64
}

// GinConfig holds the configuration for custom gin web server options
type GinConfig struct {
	ForwardedByClientIP bool
	TrustedPlatform     string
	RemoteIPHeaders     []string
	TrustedProxies      []string
}

// PubkeyConfig will map the public key configuration
type PubkeyConfig struct {
	Length int
	Type   string
	Hrp    string
}
