package nativeAuth

import "time"

// SetTimeHandler -
func (server *maxTTLAuthServer) SetTimeHandler(handler func() time.Time) {
	server.getTimeHandler = handler
}
