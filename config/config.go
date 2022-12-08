package config

import "github.com/ElrondNetwork/elrond-go-storage/storageUnit"

// Configs is a holder for the relayer configuration parameters
type Configs struct {
	GeneralConfig   Config
	ApiRoutesConfig ApiRoutesConfig
	FlagsConfig     ContextFlagsConfig
}

// Config general configuration struct
type Config struct {
	Guardian         GuardianConfig
	Proxy            ProxyConfig
	Logs             LogsConfig
	Antiflood        AntifloodConfig
	NativeAuthServer NativeAuthServerConfig
	ServiceResolver  ServiceResolverConfig
	OTP              StorageConfig
	Users            StorageConfig
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

// NativeAuthServerConfig will hold all native authentification server parameters
type NativeAuthServerConfig struct {
	Enabled       bool
	AcceptedHosts []string
}

// ApiRoutesConfig holds the configuration related to Rest API routes
type ApiRoutesConfig struct {
	Logging     ApiLoggingConfig
	APIPackages map[string]APIPackageConfig
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
	PrivateKeyFile       string
	RequestTimeInSeconds int
}

// ProxyConfig will hold settings related to the Elrond Proxy
type ProxyConfig struct {
	NetworkAddress               string
	ProxyCacherExpirationSeconds uint64
	ProxyRestAPIEntityType       string
	ProxyMaxNoncesDelta          int
	ProxyFinalityCheck           bool
}

// LogsConfig will hold settings related to the logging sub-system
type LogsConfig struct {
	LogFileLifeSpanInSec int
}

// ServiceResolverConfig will hold settings related to the service resolver
type ServiceResolverConfig struct {
	RequestTimeInSeconds uint64
}
