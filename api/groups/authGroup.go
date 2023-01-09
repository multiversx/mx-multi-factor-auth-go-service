package groups

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	elrondApiShared "github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/core"
	"github.com/ElrondNetwork/elrond-sdk-erdgo/data"
	mfaMiddleware "github.com/ElrondNetwork/multi-factor-auth-go-service/api/middleware"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/api/shared"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
	"github.com/gin-gonic/gin"
)

const (
	signTransaction          = "/sign-transaction"
	signMultipleTransactions = "/sign-multiple-transactions"
	registerPath             = "/register"
	verifyCodePath           = "/verify-code"
)

type authGroup struct {
	*baseGroup
	facade    shared.FacadeHandler
	mutFacade sync.RWMutex
}

// NewAuthGroup returns a new instance of authGroup
func NewAuthGroup(facade shared.FacadeHandler) (*authGroup, error) {
	if check.IfNil(facade) {
		return nil, fmt.Errorf("%w for node group", errors.ErrNilFacadeHandler)
	}

	ag := &authGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*elrondApiShared.EndpointHandlerData{
		{
			Path:    signTransaction,
			Method:  http.MethodPost,
			Handler: ag.signTransaction,
		},
		{
			Path:    signMultipleTransactions,
			Method:  http.MethodPost,
			Handler: ag.signMultipleTransactions,
		},
		{
			Path:    registerPath,
			Method:  http.MethodPost,
			Handler: ag.register,
		},
		{
			Path:    verifyCodePath,
			Method:  http.MethodPost,
			Handler: ag.verifyCode,
		},
	}
	ag.endpoints = endpoints

	return ag, nil
}

// signTransaction returns the transaction signed by the guardian if the verification passed
func (ag *authGroup) signTransaction(c *gin.Context) {
	var request requests.SignTransaction

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	marshalledTx := make([]byte, 0)
	if err == nil {
		userAddress := ag.extractAddressContext(c)
		marshalledTx, err = ag.facade.SignTransaction(userAddress, request)
	}
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			elrondApiShared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", ErrValidation.Error(), err.Error()),
				Code:  elrondApiShared.ReturnCodeInternalError,
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		elrondApiShared.GenericAPIResponse{
			Data:  marshalledTx,
			Error: "",
			Code:  elrondApiShared.ReturnCodeSuccess,
		},
	)
}

// signMultipleTransactions returns the transactions signed by the guardian if the verification passed
func (ag *authGroup) signMultipleTransactions(c *gin.Context) {
	var request requests.SignMultipleTransactions

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	marshalledTxs := make([][]byte, 0)
	if err == nil {
		userAddress := ag.extractAddressContext(c)
		marshalledTxs, err = ag.facade.SignMultipleTransactions(userAddress, request)
	}
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			elrondApiShared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", ErrValidation.Error(), err.Error()),
				Code:  elrondApiShared.ReturnCodeInternalError,
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		elrondApiShared.GenericAPIResponse{
			Data:  marshalledTxs,
			Error: "",
			Code:  elrondApiShared.ReturnCodeSuccess,
		},
	)
}

// register will register the user and (optionally) returns some information required
// for the user to set up the OTP on his end (eg: QR code).
func (ag *authGroup) register(c *gin.Context) {
	var request requests.RegistrationPayload

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	retData := &requests.RegisterReturnData{}
	if err == nil {
		userAddress := ag.extractAddressContext(c)
		retData.QR, retData.GuardianAddress, err = ag.facade.RegisterUser(userAddress, request)
	}
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			elrondApiShared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", ErrRegister.Error(), err.Error()),
				Code:  elrondApiShared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		elrondApiShared.GenericAPIResponse{
			Data:  retData,
			Error: "",
			Code:  elrondApiShared.ReturnCodeSuccess,
		},
	)
}

// verifyCode validates a code
func (ag *authGroup) verifyCode(c *gin.Context) {
	var request requests.VerificationPayload

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	if err == nil {
		userAddress := ag.extractAddressContext(c)
		err = ag.facade.VerifyCode(userAddress, request)
	}
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			elrondApiShared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", ErrValidation.Error(), err.Error()),
				Code:  elrondApiShared.ReturnCodeInternalError,
			},
		)
		return
	}
	c.JSON(
		http.StatusOK,
		elrondApiShared.GenericAPIResponse{
			Data:  "",
			Error: "",
			Code:  elrondApiShared.ReturnCodeSuccess,
		},
	)
}

// UpdateFacade will update the facade
func (ag *authGroup) UpdateFacade(newFacade shared.FacadeHandler) error {
	if check.IfNil(newFacade) {
		return errors.ErrNilFacadeHandler
	}

	ag.mutFacade.Lock()
	ag.facade = newFacade
	ag.mutFacade.Unlock()

	return nil
}

func (ag *authGroup) extractAddressContext(c *gin.Context) core.AddressHandler {
	userAddressStr := c.GetString(mfaMiddleware.UserAddressKey)
	userAddress, _ := data.NewAddressFromBech32String(userAddressStr)
	return userAddress
}

// IsInterfaceNil returns true if there is no value under the interface
func (ag *authGroup) IsInterfaceNil() bool {
	return ag == nil
}
