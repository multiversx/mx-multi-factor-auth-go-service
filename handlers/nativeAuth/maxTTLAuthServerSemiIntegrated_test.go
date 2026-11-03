package nativeAuth_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/multiversx/mx-chain-core-go/core/pubkeyConverter"
	"github.com/multiversx/mx-chain-crypto-go/signing"
	"github.com/multiversx/mx-chain-crypto-go/signing/ed25519"
	"github.com/multiversx/mx-chain-storage-go/lrucache"
	"github.com/multiversx/mx-sdk-go/authentication"
	"github.com/multiversx/mx-sdk-go/authentication/native"
	"github.com/multiversx/mx-sdk-go/blockchain/cryptoProvider"
	"github.com/multiversx/mx-sdk-go/data"
	"github.com/multiversx/mx-sdk-go/examples"
	"github.com/multiversx/mx-sdk-go/interactors"
	"github.com/multiversx/mx-sdk-go/testsCommon"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-multi-factor-auth-go-service/handlers/nativeAuth"
)

func createSignedToken(t *testing.T, ttl int64, blockTimestamp int64) string {
	keyGen := signing.NewKeyGenerator(ed25519.NewEd25519())
	privateKeyBytes, err := interactors.NewWallet().LoadPrivateKeyFromPemData([]byte(examples.AlicePemContents))
	require.Nil(t, err)
	cryptoCompHolder, err := cryptoProvider.NewCryptoComponentsHolder(keyGen, privateKeyBytes)
	require.Nil(t, err)

	client, err := native.NewNativeAuthClient(native.ArgsNativeAuthClient{
		Signer:    cryptoProvider.NewSigner(),
		ExtraInfo: struct{}{},
		Proxy: &testsCommon.ProxyStub{
			GetHyperBlockByNonceCalled: func(ctx context.Context, nonce uint64) (*data.HyperBlock, error) {
				return &data.HyperBlock{Timestamp: uint64(blockTimestamp), Hash: blockHash}, nil
			},
		},
		CryptoComponentsHolder: cryptoCompHolder,
		TokenExpiryInSeconds:   ttl,
		TokenHandler:           native.NewAuthTokenHandler(),
		Host:                   "host",
	})
	require.Nil(t, err)

	token, err := client.GetAccessToken()
	require.Nil(t, err)

	return token
}

func createRealAuthServer(t *testing.T, blockTimestamp int64) authentication.AuthServer {
	httpClient := &httpClientWrapperStub{
		getHTTPCalled: func(ctx context.Context, endpoint string) ([]byte, int, error) {
			buff, _ := json.Marshal(&data.Block{Timestamp: int(blockTimestamp)})
			return buff, http.StatusOK, nil
		},
	}
	cacher, err := lrucache.NewCache(10)
	require.Nil(t, err)
	converter, err := pubkeyConverter.NewBech32PubkeyConverter(32, "erd")
	require.Nil(t, err)

	sdkServer, err := native.NewNativeAuthServer(native.ArgsNativeAuthServer{
		HttpClientWrapper: httpClient,
		TokenHandler:      native.NewAuthTokenHandler(),
		Signer:            cryptoProvider.NewSigner(),
		PubKeyConverter:   converter,
		KeyGenerator:      signing.NewKeyGenerator(ed25519.NewEd25519()),
		TimestampsCacher:  cacher,
	})
	require.Nil(t, err)

	server, err := nativeAuth.NewMaxTTLAuthServer(nativeAuth.ArgsMaxTTLAuthServer{
		AuthServer:        sdkServer,
		HttpClientWrapper: httpClient,
		TimestampsCacher:  cacher,
		MaxTTLInSec:       maxTTL,
	})
	require.Nil(t, err)

	return server
}

func TestMaxTTLAuthServer_SignedTokens(t *testing.T) {
	t.Parallel()

	oneYear := int64(365 * 24 * 3600)
	tokenHandler := native.NewAuthTokenHandler()

	t.Run("long ttl token should be accepted within max ttl", func(t *testing.T) {
		t.Parallel()

		timestamp := time.Now().Unix() - 10
		token, err := tokenHandler.Decode(createSignedToken(t, oneYear, timestamp))
		require.Nil(t, err)
		require.Nil(t, createRealAuthServer(t, timestamp).Validate(token))
	})
	t.Run("long ttl token should be rejected after max ttl", func(t *testing.T) {
		t.Parallel()

		timestamp := time.Now().Unix() - maxTTL - 10
		token, err := tokenHandler.Decode(createSignedToken(t, oneYear, timestamp))
		require.Nil(t, err)
		err = createRealAuthServer(t, timestamp).Validate(token)
		require.ErrorIs(t, err, authentication.ErrTokenExpired)
	})
	t.Run("short ttl token should keep its own expiry", func(t *testing.T) {
		t.Parallel()

		timestamp := time.Now().Unix() - 70
		token, err := tokenHandler.Decode(createSignedToken(t, 60, timestamp))
		require.Nil(t, err)
		err = createRealAuthServer(t, timestamp).Validate(token)
		require.ErrorIs(t, err, authentication.ErrTokenExpired)
	})
}
