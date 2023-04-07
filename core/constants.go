package core

// WebServerOffString represents the constant used to switch off the web server
const WebServerOffString = "off"

// DBType represents the type of DB
type DBType string

// LevelDB is the local levelDB
const LevelDB DBType = "levelDB"

const (
	getAccountEndpointFormat      = "address/%s"
	getGuardianDataEndpointFormat = "address/%s/guardian-data"
)
