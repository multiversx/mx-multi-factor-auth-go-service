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
	sendTransaction = "/sendTransaction"
	registerPath    = "/register"
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
			Path:    sendTransaction,
			Method:  http.MethodPost,
			Handler: ag.sendTransaction,
		},
		{
			Path:    registerPath,
			Method:  http.MethodPost,
			Handler: ag.register,
		},
	}
	ag.endpoints = endpoints

	return ag, nil
}

// sendTransaction returns will send the transaction signed by the guardian if the verification passed
func (ag *authGroup) sendTransaction(c *gin.Context) {
	var request requests.SendTransaction

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	hash := ""
	if err == nil {
		hash, err = ag.facade.Validate(request)
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
			Data:  hash,
			Error: "",
			Code:  elrondApiShared.ReturnCodeSuccess,
		},
	)
}

// register will register a new provider for the user
// and (optionally) returns some information required for the user to set up the OTP on his end (eg: QR code).
func (ag *authGroup) register(c *gin.Context) {
	var request requests.Register

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
