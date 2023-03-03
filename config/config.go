package config

import "github.com/multiversx/mx-chain-storage-go/storageUnit"

// Configs is a holder for the relayer configuration parameters
type Configs struct {
	GeneralConfig   Config
	ApiRoutesConfig ApiRoutesConfig
	FlagsConfig     ContextFlagsConfig
}

// Config general configuration struct
type Config struct {
	Guardian        GuardianConfig
	General         GeneralConfig
	Proxy           ProxyConfig
	Api             ApiConfig
	Logs            LogsConfig
	Antiflood       AntifloodConfig
	ServiceResolver ServiceResolverConfig
	ShardedStorage  ShardedStorageConfig
	Buckets         BucketsConfig
	TwoFactor       TwoFactorConfig
}

// ShardedStorageConfig is the configuration for the sharded storage
type ShardedStorageConfig struct {
	LocalStorageEnabled bool
	Users               StorageConfig
}

// StorageConfig will map the storage unit configuration
type StorageConfig struct {
	Cache storageUnit.CacheConfig
	DB    storageUnit.DBConfig
}

// ContextFlagsConfig the configuration for flags
type ContextFlagsConfig struct {
	WorkingDir           string
	LogLevel             string
	ConfigurationFile    string
	ConfigurationApiFile string
	RestApiInterface     string
	DisableAnsiColor     bool
	SaveLogFile          bool
	EnableLogName        bool
	EnablePprof          bool
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
	Name string
	Open bool
}

// GuardianConfig holds the configuration for the guardian
type GuardianConfig struct {
	MnemonicFile         string
	RequestTimeInSeconds int
}

// GeneralConfig holds the general configuration for the service
type GeneralConfig struct {
	Marshalizer string
}

// ProxyConfig will hold settings related to the Proxy
type ProxyConfig struct {
	NetworkAddress               string
	ProxyCacherExpirationSeconds uint64
	ProxyRestAPIEntityType       string
	ProxyMaxNoncesDelta          int
	ProxyFinalityCheck           bool
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
	RequestTimeInSeconds uint64
	SkipTxUserSigVerify  bool
}

// BucketsConfig will hold settings related to buckets
type BucketsConfig struct {
	NumberOfBuckets uint32
}

// TwoFactorConfig will hold settings related to the two factor totp
type TwoFactorConfig struct {
	Issuer string
	Digits int
}
