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
	"github.com/gin-gonic/gin"
)

const (
	validatePath = "/validate"
	registerPath = "/register"
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

	ng := &authGroup{
		facade:    facade,
		baseGroup: &baseGroup{},
	}

	endpoints := []*elrondApiShared.EndpointHandlerData{
		{
			Path:    validatePath,
			Method:  http.MethodPost,
			Handler: ng.validate,
		},
		{
			Path:    registerPath,
			Method:  http.MethodPost,
			Handler: ng.register,
		},
	}
	ng.endpoints = endpoints

	return ng, nil
}

// getCollectionRarity returns the information of a provided metric
func (ng *authGroup) validate(c *gin.Context) {
	var guardianValidateRequest GuardianValidateRequest

	err := json.NewDecoder(c.Request.Body).Decode(&guardianValidateRequest)
	isValid := false
	if err == nil {
		isValid, err = ng.facade.Validate(guardianValidateRequest)
	}
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			elrondApiShared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", ErrComputingRarity.Error(), err.Error()),
				Code:  elrondApiShared.ReturnCodeInternalError,
			},
		)
		return
	}
	jsonString := fmt.Sprintf("{\"valid\":%t}", isValid)
	var data interface{}
	err = json.Unmarshal([]byte(jsonString), data)
	if err != nil {
		return
	}
	c.JSON(
		http.StatusOK,
		elrondApiShared.GenericAPIResponse{
			Data:  data,
			Error: "",
			Code:  elrondApiShared.ReturnCodeSuccess,
		},
	)
}

func (ng *authGroup) register(c *gin.Context) {
	var guardianRegisterRequest GuardianRegisterRequest

	err := json.NewDecoder(c.Request.Body).Decode(&guardianRegisterRequest)
	var qr []byte
	if err == nil {
		qr, err = ng.facade.RegisterUser(guardianRegisterRequest)
	}
	if err != nil {
		c.JSON(
			http.StatusInternalServerError,
			elrondApiShared.GenericAPIResponse{
				Data:  nil,
				Error: fmt.Sprintf("%s: %s", ErrComputingRarity.Error(), err.Error()),
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
func (ng *authGroup) UpdateFacade(newFacade shared.FacadeHandler) error {
	if check.IfNil(newFacade) {
		return errors.ErrNilFacadeHandler
	}

	ng.mutFacade.Lock()
	ng.facade = newFacade
	ng.mutFacade.Unlock()

	return nil
}

// IsInterfaceNil returns true if there is no value under the interface
func (ng *authGroup) IsInterfaceNil() bool {
	return ng == nil
}
