package nativeAuth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-sdk-go/authentication"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	blockByHashEndpoint = "blocks/%s"
	int64Size           = 64
)

// ArgsMaxTTLAuthServer is the DTO used to create a new instance of maxTTLAuthServer
type ArgsMaxTTLAuthServer struct {
	AuthServer        authentication.AuthServer
	HttpClientWrapper authentication.HttpClientWrapper
	TimestampsCacher  TimestampsCacher
	MaxTTLInSec       int64
}

type maxTTLAuthServer struct {
	authServer        authentication.AuthServer
	httpClientWrapper authentication.HttpClientWrapper
	cacher            TimestampsCacher
	maxTTLInSec       int64
	getTimeHandler    func() time.Time
}

// NewMaxTTLAuthServer returns an auth server which caps the tokens' ttl to the configured max ttl
func NewMaxTTLAuthServer(args ArgsMaxTTLAuthServer) (*maxTTLAuthServer, error) {
	err := checkArgs(args)
	if err != nil {
		return nil, err
	}

	return &maxTTLAuthServer{
		authServer:        args.AuthServer,
		httpClientWrapper: args.HttpClientWrapper,
		cacher:            args.TimestampsCacher,
		maxTTLInSec:       args.MaxTTLInSec,
		getTimeHandler:    time.Now,
	}, nil
}

func checkArgs(args ArgsMaxTTLAuthServer) error {
	if check.IfNil(args.AuthServer) {
		return ErrNilAuthServer
	}
	if check.IfNil(args.HttpClientWrapper) {
		return authentication.ErrNilHttpClientWrapper
	}
	if check.IfNil(args.TimestampsCacher) {
		return authentication.ErrNilCacher
	}
	if args.MaxTTLInSec <= 0 {
		return fmt.Errorf("%w for MaxTTLInSec, received %d", ErrInvalidValue, args.MaxTTLInSec)
	}

	return nil
}

// Validate validates the token with the wrapped auth server, then checks it against the configured max ttl
func (server *maxTTLAuthServer) Validate(authToken authentication.AuthToken) error {
	err := server.authServer.Validate(authToken)
	if err != nil {
		return err
	}

	if authToken.GetTtl() <= server.maxTTLInSec {
		return nil
	}

	timestamp, err := server.getBlockTimestamp(authToken.GetBlockHash())
	if err != nil {
		return err
	}

	expireTime := time.Unix(timestamp+server.maxTTLInSec, 0)
	if server.getTimeHandler().After(expireTime) {
		return fmt.Errorf("%w, max ttl of %d seconds exceeded", authentication.ErrTokenExpired, server.maxTTLInSec)
	}

	return nil
}

func (server *maxTTLAuthServer) getBlockTimestamp(blockHash string) (int64, error) {
	cachedTimestampValue, found := server.cacher.Get([]byte(blockHash))
	if found {
		timestamp, ok := cachedTimestampValue.(int64)
		if !ok {
			return 0, fmt.Errorf("%w while casting timestamp value: %v", authentication.ErrInvalidValue, cachedTimestampValue)
		}

		return timestamp, nil
	}

	buff, code, err := server.httpClientWrapper.GetHTTP(context.Background(), fmt.Sprintf(blockByHashEndpoint, blockHash))
	if err != nil || code != http.StatusOK {
		return 0, authentication.CreateHTTPStatusError(code, err)
	}

	var block data.Block
	err = json.Unmarshal(buff, &block)
	if err != nil {
		return 0, err
	}

	timestamp := int64(block.Timestamp)
	server.cacher.Put([]byte(blockHash), timestamp, int64Size)

	return timestamp, nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (server *maxTTLAuthServer) IsInterfaceNil() bool {
	return server == nil
}
