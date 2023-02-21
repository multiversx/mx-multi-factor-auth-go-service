package groups

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	mfaMiddleware "github.com/multiversx/multi-factor-auth-go-service/api/middleware"
	"github.com/multiversx/multi-factor-auth-go-service/api/shared"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-go/api/errors"
	chainApiShared "github.com/multiversx/mx-chain-go/api/shared"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	signTransactionPath          = "/sign-transaction"
	signMultipleTransactionsPath = "/sign-multiple-transactions"
	registerPath                 = "/register"
	verifyCodePath               = "/verify-code"
)

var guardianLog = logger.GetOrCreate("guardianGroup")

type guardianGroup struct {
	*baseGroup
	facade    shared.FacadeHandler
	mutFacade sync.RWMutex
}

// NewGuardianGroup returns a new instance of guardianGroup
func NewGuardianGroup(facade shared.FacadeHandler) (*guardianGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for node group", errors.ErrNilFacadeHandler)
	}

	gg := &guardianGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*chainApiShared.EndpointHandlerData{
		{
			Path:    signTransactionPath,
			Method:  http.MethodPost,
			Handler: gg.signTransaction,
		},
		{
			Path:    signMultipleTransactionsPath,
			Method:  http.MethodPost,
			Handler: gg.signMultipleTransactions,
		},
		{
			Path:    registerPath,
			Method:  http.MethodPost,
			Handler: gg.register,
		},
		{
			Path:    verifyCodePath,
			Method:  http.MethodPost,
			Handler: gg.verifyCode,
		},
	}
	gg.endpoints = endpoints

	return gg, nil
}

// signTransaction returns the transaction signed by the guardian if the verification passed
func (gg *guardianGroup) signTransaction(c *gin.Context) {
	var request requests.SignTransaction

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	marshalledTx := make([]byte, 0)
	var signTransactionResponse *requests.SignTransactionResponse
	if err == nil {
		userAddress := gg.extractAddressContext(c)
		marshalledTx, err = gg.facade.SignTransaction(userAddress, request)
		if err == nil {
			signTransactionResponse, err = createSignTransactionResponse(marshalledTx)
		}
	}
	if err != nil {
		guardianLog.Trace("cannot sign transaction", "error", err.Error(), "transaction", request.Tx)
		c.JSON(
			http.StatusInternalServerError,
			chainApiShared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  chainApiShared.ReturnCodeInternalError,
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		chainApiShared.GenericAPIResponse{
			Data:  signTransactionResponse,
			Error: "",
			Code:  chainApiShared.ReturnCodeSuccess,
		},
	)
}

// signMultipleTransactions returns the transactions signed by the guardian if the verification passed
func (gg *guardianGroup) signMultipleTransactions(c *gin.Context) {
	var request requests.SignMultipleTransactions

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	marshalledTxs := make([][]byte, 0)
	signMultipleTransactionsResponse := &requests.SignMultipleTransactionsResponse{}
	if err == nil {
		userAddress := gg.extractAddressContext(c)
		marshalledTxs, err = gg.facade.SignMultipleTransactions(userAddress, request)
		if err == nil {
			signMultipleTransactionsResponse, err = createSignMultipleTransactionsResponse(marshalledTxs)
		}
	}
	if err != nil {
		guardianLog.Trace("cannot sign transactions", "error", err.Error(), "transactions", request.Txs)
		c.JSON(
			http.StatusInternalServerError,
			chainApiShared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  chainApiShared.ReturnCodeInternalError,
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		chainApiShared.GenericAPIResponse{
			Data:  signMultipleTransactionsResponse,
			Error: "",
			Code:  chainApiShared.ReturnCodeSuccess,
		},
	)
}

func createSignMultipleTransactionsResponse(marshalledTxs [][]byte) (*requests.SignMultipleTransactionsResponse, error) {
	signMultipleTransactionsResponse := &requests.SignMultipleTransactionsResponse{
		Txs: make([]data.Transaction, 0),
	}
	for _, marshalledTx := range marshalledTxs {
		unmarshalledTx := data.Transaction{}
		err := json.Unmarshal(marshalledTx, &unmarshalledTx)
		if err != nil {
			return nil, err
		}
		signMultipleTransactionsResponse.Txs = append(signMultipleTransactionsResponse.Txs, unmarshalledTx)
	}

	return signMultipleTransactionsResponse, nil
}

func createSignTransactionResponse(marshalledTx []byte) (*requests.SignTransactionResponse, error) {
	signTransactionResponse := &requests.SignTransactionResponse{}
	err := json.Unmarshal(marshalledTx, &signTransactionResponse.Tx)
	if err != nil {
		return nil, err
	}

	return signTransactionResponse, nil
}

// register will register the user and (optionally) returns some information required
// for the user to set up the OTP on his end (eg: QR code).
func (gg *guardianGroup) register(c *gin.Context) {
	var request requests.RegistrationPayload

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	retData := &requests.RegisterReturnData{}
	if err == nil {
		userAddress := gg.extractAddressContext(c)
		retData.QR, retData.GuardianAddress, err = gg.facade.RegisterUser(userAddress, request)
	}
	if err != nil {
		guardianLog.Trace("cannot register", "error", err.Error())
		c.JSON(
			http.StatusInternalServerError,
			chainApiShared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  chainApiShared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		chainApiShared.GenericAPIResponse{
			Data:  retData,
			Error: "",
			Code:  chainApiShared.ReturnCodeSuccess,
		},
	)
}

// verifyCode validates a code
func (gg *guardianGroup) verifyCode(c *gin.Context) {
	var request requests.VerificationPayload

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	if err == nil {
		userAddress := gg.extractAddressContext(c)
		err = gg.facade.VerifyCode(userAddress, request)
	}
	if err != nil {
		guardianLog.Trace("cannot verify guardian", "error", err.Error())
		c.JSON(
			http.StatusInternalServerError,
			chainApiShared.GenericAPIResponse{
				Data:  nil,
				Error: err.Error(),
				Code:  chainApiShared.ReturnCodeInternalError,
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		chainApiShared.GenericAPIResponse{
			Data:  "",
			Error: "",
			Code:  chainApiShared.ReturnCodeSuccess,
		},
	)
}

// UpdateFacade will update the facade
func (gg *guardianGroup) UpdateFacade(newFacade shared.FacadeHandler) error {
	if check.IfNil(newFacade) {
		return errors.ErrNilFacadeHandler
	}

	gg.mutFacade.Lock()
	gg.facade = newFacade
	gg.mutFacade.Unlock()

	return nil
}

func (gg *guardianGroup) extractAddressContext(c *gin.Context) core.AddressHandler {
	userAddressStr := c.GetString(mfaMiddleware.UserAddressKey)
	userAddress, _ := data.NewAddressFromBech32String(userAddressStr)
	return userAddress
}

// IsInterfaceNil returns true if there is no value under the interface
func (gg *guardianGroup) IsInterfaceNil() bool {
	return gg == nil
}
