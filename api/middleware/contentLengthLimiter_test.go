package middleware

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	"github.com/multiversx/mx-chain-go/api/shared"
	"github.com/stretchr/testify/require"

	"github.com/multiversx/mx-multi-factor-auth-go-service/config"
	"github.com/multiversx/mx-multi-factor-auth-go-service/core/requests"
)

func startServerWithContentLength(providedMap map[string]config.APIPackageConfig, handler func(c *gin.Context)) (*gin.Engine, error) {
	ws := gin.New()
	ws.Use(cors.Default())

	lengthLimiterMiddleware, err := NewContentLengthLimiter(providedMap)
	if err != nil {
		return nil, err
	}

	if check.IfNil(lengthLimiterMiddleware) {
		return nil, errors.New("length limiter middleware cannot be nil")
	}

	ws.Use(lengthLimiterMiddleware.MiddlewareHandlerFunc())

	ginAddressRoutes := ws.Group("/guardian")

	ginAddressRoutes.Handle(http.MethodPost, "/sign-message", handler)
	ginAddressRoutes.Handle(http.MethodPost, "/sign-transaction", handler)
	ginAddressRoutes.Handle(http.MethodPost, "/register", handler)

	return ws, nil
}

func TestContentLengthLimiter(t *testing.T) {
	t.Parallel()

	providedMap := map[string]config.APIPackageConfig{
		"guardian": {
			Routes: []config.RouteConfig{
				{
					Name:             "/register",
					Open:             true,
					Auth:             true,
					MaxContentLength: 100,
				},
				{
					Name:             "/sign-transaction",
					Open:             true,
					Auth:             false,
					MaxContentLength: 400,
				},
				{
					Name:             "/sign-message",
					Open:             true,
					Auth:             false,
					MaxContentLength: 100,
				},
			},
		},
	}

	t.Run("content too large", func(t *testing.T) {
		t.Parallel()

		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}
		ws, err := startServerWithContentLength(providedMap, handlerFunc)
		require.NoError(t, err)

		registrationPayload := requests.SignMessage{
			Code:         "123456",
			SecondCode:   "654321",
			Message:      "b1c3ce06-5ac0-4244-bb5d-b14ff9563fdc",
			UserAddr:     "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th", // Alice
			GuardianAddr: "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx",
		}
		body, err := json.Marshal(registrationPayload)
		require.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPost, "/guardian/sign-message", bytes.NewReader(body))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		require.Equal(t, http.StatusRequestEntityTooLarge, resp.Code)
		require.Nil(t, response.Data)
		require.Equal(t, shared.ReturnCodeRequestError, response.Code)
		require.Contains(t, response.Error, ErrContentLengthTooLarge.Error())
		require.Contains(t, response.Error, "cannot process request")
	})

	t.Run("unknown content length", func(t *testing.T) {
		t.Parallel()

		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}

		ws, err := startServerWithContentLength(providedMap, handlerFunc)
		require.NoError(t, err)

		registrationPayload := requests.RegistrationPayload{}
		body, err := json.Marshal(registrationPayload)
		require.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPost, "/guardian/register", bytes.NewReader(body))
		resp := httptest.NewRecorder()

		req.ContentLength = -1
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		require.Equal(t, http.StatusBadRequest, resp.Code)
		require.Nil(t, response.Data)
		require.Equal(t, shared.ReturnCodeRequestError, response.Code)
		require.Contains(t, response.Error, ErrUnknownContentLength.Error())
		require.Contains(t, response.Error, "cannot process request")
	})

	t.Run("content max size, too small", func(t *testing.T) {
		t.Parallel()

		handlerFunc := func(c *gin.Context) {
			c.JSON(200, "ok")
		}

		contentLength := uint64(9)

		faultyConfig := map[string]config.APIPackageConfig{
			"guardian": {
				Routes: []config.RouteConfig{
					{
						Name:             "/register",
						Open:             true,
						Auth:             true,
						MaxContentLength: contentLength,
					},
				},
			},
		}

		ws, err := startServerWithContentLength(faultyConfig, handlerFunc)
		require.Nil(t, ws)
		require.Equal(t, fmt.Errorf("%w, min expected %d, received %d", ErrMaxContentLengthTooSmall, minSizeBytes,
			contentLength), err)

	})

	t.Run("should work", func(t *testing.T) {
		t.Parallel()

		handlerFunc := func(c *gin.Context) {
			resp := requests.SignTransactionResponse{
				Tx: transaction.FrontendTransaction{},
			}
			c.JSON(http.StatusOK, shared.GenericAPIResponse{
				Data:  resp,
				Error: "",
				Code:  shared.ReturnCodeSuccess,
			})
		}
		ws, err := startServerWithContentLength(providedMap, handlerFunc)
		require.NoError(t, err)

		registrationPayload := requests.SignTransaction{
			Code:       "123456",
			SecondCode: "654321",
			Tx: transaction.FrontendTransaction{
				Sender:       "erd1qyu5wthldzr8wx5c9ucg8kjagg0jfs53s8nr3zpz3hypefsdd8ssycr6th",
				Signature:    hex.EncodeToString([]byte("signature")),
				GuardianAddr: "erd1spyavw0956vq68xj8y4tenjpq2wd5a9p2c6j8gsz7ztyrnpxrruqzu66jx",
			},
		}
		body, err := json.Marshal(registrationPayload)
		require.NoError(t, err)

		req, _ := http.NewRequest(http.MethodPost, "/guardian/sign-transaction", bytes.NewReader(body))
		resp := httptest.NewRecorder()
		ws.ServeHTTP(resp, req)

		var response shared.GenericAPIResponse

		_ = json.NewDecoder(resp.Body).Decode(&response)
		require.Equal(t, http.StatusOK, resp.Code)
		require.NotNil(t, response.Data)
		require.Equal(t, shared.ReturnCodeSuccess, response.Code)
		require.Empty(t, response.Error)
	})
}
