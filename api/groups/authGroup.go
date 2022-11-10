package groups

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/elrond-go/api/errors"
	elrondApiShared "github.com/ElrondNetwork/elrond-go/api/shared"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/api/shared"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core/requests"
	"github.com/gin-gonic/gin"
)

const (
	signTransaction          = "/sign-transaction"
	signMultipleTransactions = "/sign-multiple-transactions"
	registerPath             = "/register"
	getGuardianAddressPath   = "/generate-guardian"
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
			Path:    getGuardianAddressPath,
			Method:  http.MethodPost,
			Handler: ag.getGuardianAddress,
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
		marshalledTx, err = ag.facade.SignTransaction(request)
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
	mashalledTxs := make([][]byte, 0)
	if err == nil {
		mashalledTxs, err = ag.facade.SignMultipleTransactions(request)
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
			Data:  mashalledTxs,
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
	var qr []byte
	if err == nil {
		qr, err = ag.facade.RegisterUser(request)
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
			Data:  qr,
			Error: "",
			Code:  elrondApiShared.ReturnCodeSuccess,
		},
	)
}

// getGuardianAddress will return a unique address of a guardian
func (ag *authGroup) getGuardianAddress(c *gin.Context) {
	var request requests.GetGuardianAddress

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	var guardianAddress string
	if err == nil {
		guardianAddress, err = ag.facade.GetGuardianAddress(request)
	}
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			elrondApiShared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", ErrGetGuardianAddress.Error(), err.Error()),
				Code:  elrondApiShared.ReturnCodeInternalError,
			},
		)
		return
	}

	c.JSON(
		http.StatusOK,
		elrondApiShared.GenericAPIResponse{
			Data:  guardianAddress,
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
		err = ag.facade.VerifyCode(request)
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

// IsInterfaceNil returns true if there is no value under the interface
func (ag *authGroup) IsInterfaceNil() bool {
	return ag == nil
}
