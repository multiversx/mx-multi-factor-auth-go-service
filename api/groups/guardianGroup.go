package groups

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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
	registeredUsersPath          = "/registered-users"
	tcsConfig                    = "/config"

	tokensMismatchError = "Tokens mismatch"
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
		{
			Path:    registeredUsersPath,
			Method:  http.MethodGet,
			Handler: gg.registeredUsers,
		},
		{
			Path:    tcsConfig,
			Method:  http.MethodGet,
			Handler: gg.config,
		},
	}
	gg.endpoints = endpoints

	return gg, nil
}

// signTransaction returns the transaction signed by the guardian if the verification passed
func (gg *guardianGroup) signTransaction(c *gin.Context) {
	var request requests.SignTransaction
	var debugErr error

	userIp := c.GetString(mfaMiddleware.UserIpKey)
	userAgent := c.GetString(mfaMiddleware.UserAgentKey)
	defer func() {
		logArgs := []interface{}{
			"route", signTransactionPath,
			"ip", userIp,
			"user agent", userAgent,
			"transaction", getPrintableTxData(&request.Tx),
		}
		if debugErr == nil {
			logArgs = append(logArgs, "result", "success")
		} else {
			if strings.Contains(debugErr.Error(), tokensMismatchError) {
				logArgs = append(logArgs, "code", request.Code)
			}
			logArgs = append(logArgs, "error", debugErr.Error())
		}

		log.Info("Request info", logArgs...)
	}()

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	if err != nil {
		debugErr = fmt.Errorf("%w while decoding request", err)
		guardianLog.Debug("cannot decode sign transaction request",
			"error", err.Error())
		returnStatus(c, nil, http.StatusBadRequest, err.Error(), chainApiShared.ReturnCodeRequestError)
		return
	}

	var signTransactionResponse *requests.SignTransactionResponse
	marshalledTx, err := gg.facade.SignTransaction(userIp, request)
	if err != nil {
		debugErr = fmt.Errorf("%w while signing transaction", err)
		guardianLog.Debug("cannot sign transaction",
			"ip", userIp,
			"user agent", userAgent,
			"error", err.Error())
		returnStatus(c, nil, http.StatusInternalServerError, err.Error(), chainApiShared.ReturnCodeInternalError)
		return
	}

	signTransactionResponse, err = createSignTransactionResponse(marshalledTx)
	if err != nil {
		debugErr = fmt.Errorf("%w while creating response", err)
		guardianLog.Debug("cannot create sign transaction response",
			"ip", userIp,
			"user agent", userAgent,
			"error", err.Error())
		returnStatus(c, nil, http.StatusInternalServerError, err.Error(), chainApiShared.ReturnCodeInternalError)
		return
	}

	returnStatus(c, signTransactionResponse, http.StatusOK, "", chainApiShared.ReturnCodeSuccess)
}

// signMultipleTransactions returns the transactions signed by the guardian if the verification passed
func (gg *guardianGroup) signMultipleTransactions(c *gin.Context) {
	var request requests.SignMultipleTransactions
	var debugErr error

	userIp := c.GetString(mfaMiddleware.UserIpKey)
	userAgent := c.GetString(mfaMiddleware.UserAgentKey)
	defer func() {
		logArgs := []interface{}{
			"route", signMultipleTransactionsPath,
			"ip", userIp,
			"user agent", userAgent,
			"transactions", getPrintableTxData(&request.Txs),
		}
		if debugErr == nil {
			logArgs = append(logArgs, "result", "success")
		} else {
			if strings.Contains(debugErr.Error(), tokensMismatchError) {
				logArgs = append(logArgs, "code", request.Code)
			}
			logArgs = append(logArgs, "error", debugErr.Error())
		}

		log.Info("Request info", logArgs...)
	}()

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	if err != nil {
		debugErr = fmt.Errorf("%w while decoding request", err)
		guardianLog.Debug("cannot decode sign transactions request",
			"error", err.Error())
		returnStatus(c, nil, http.StatusBadRequest, err.Error(), chainApiShared.ReturnCodeRequestError)
		return
	}

	marshalledTxs, err := gg.facade.SignMultipleTransactions(userIp, request)
	if err != nil {
		debugErr = fmt.Errorf("%w while signing transactions", err)
		guardianLog.Debug("cannot sign transactions",
			"ip", userIp,
			"user agent", userAgent,
			"error", err.Error())
		returnStatus(c, nil, http.StatusInternalServerError, err.Error(), chainApiShared.ReturnCodeInternalError)
		return
	}

	var signMultipleTransactionsResponse *requests.SignMultipleTransactionsResponse
	signMultipleTransactionsResponse, err = createSignMultipleTransactionsResponse(marshalledTxs)
	if err != nil {
		debugErr = fmt.Errorf("%w while creating response", err)
		guardianLog.Debug("cannot create sign transactions response",
			"ip", userIp,
			"user agent", userAgent,
			"error", err.Error())
		returnStatus(c, nil, http.StatusInternalServerError, err.Error(), chainApiShared.ReturnCodeInternalError)
		return
	}

	returnStatus(c, signMultipleTransactionsResponse, http.StatusOK, "", chainApiShared.ReturnCodeSuccess)
}

func createSignMultipleTransactionsResponse(marshalledTxs [][]byte) (*requests.SignMultipleTransactionsResponse, error) {
	signMultipleTransactionsResponse := &requests.SignMultipleTransactionsResponse{
		Txs: make([]data.Transaction, 0),
	}
	for i, marshalledTx := range marshalledTxs {
		unmarshalledTx := data.Transaction{}
		err := json.Unmarshal(marshalledTx, &unmarshalledTx)
		if err != nil {
			return nil, fmt.Errorf("%w for tx with index %d", err, i)
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
	var userAddress core.AddressHandler
	retData := &requests.RegisterReturnData{}
	var debugErr error

	userIp := c.GetString(mfaMiddleware.UserIpKey)
	userAgent := c.GetString(mfaMiddleware.UserAgentKey)
	defer func() {
		logArgs := []interface{}{
			"route", registerPath,
			"ip", userIp,
			"user agent", userAgent,
		}

		if !check.IfNil(userAddress) {
			logArgs = append(logArgs, "address", userAddress.AddressAsBech32String())
		}

		if debugErr == nil {
			logArgs = append(logArgs, "result", "success")
			logArgs = append(logArgs, "returned guardian", retData.GuardianAddress)
		} else {
			logArgs = append(logArgs, "error", debugErr.Error())
		}

		log.Info("Request info", logArgs...)
	}()

	userAddress, err := gg.extractAddressContext(c)
	if err != nil {
		debugErr = fmt.Errorf("%w while extracting user address", err)
		guardianLog.Debug("cannot extract user address for register", "error", err.Error())
		returnStatus(c, nil, http.StatusBadRequest, err.Error(), chainApiShared.ReturnCodeRequestError)
		return
	}

	var request requests.RegistrationPayload
	err = json.NewDecoder(c.Request.Body).Decode(&request)
	if err != nil {
		debugErr = fmt.Errorf("%w while decoding request", err)
		guardianLog.Debug("cannot decode register request",
			"userAddress", userAddress.AddressAsBech32String(),
			"error", err.Error())
		returnStatus(c, nil, http.StatusBadRequest, err.Error(), chainApiShared.ReturnCodeRequestError)
		return
	}

	retData.QR, retData.GuardianAddress, err = gg.facade.RegisterUser(userAddress, request)
	if err != nil {
		debugErr = fmt.Errorf("%w while registering", err)
		guardianLog.Debug("cannot register",
			"userAddress", userAddress.AddressAsBech32String(),
			"error", err.Error())
		returnStatus(c, nil, http.StatusInternalServerError, err.Error(), chainApiShared.ReturnCodeInternalError)
		return
	}

	returnStatus(c, retData, http.StatusOK, "", chainApiShared.ReturnCodeSuccess)
}

// verifyCode validates a code
func (gg *guardianGroup) verifyCode(c *gin.Context) {
	var request requests.VerificationPayload
	var userAddress core.AddressHandler
	var debugErr error

	userIp := c.GetString(mfaMiddleware.UserIpKey)
	userAgent := c.GetString(mfaMiddleware.UserAgentKey)
	defer func() {
		logArgs := []interface{}{
			"route", verifyCodePath,
			"ip", userIp,
			"user agent", userAgent,
			"guardian", request.Guardian,
		}

		if !check.IfNil(userAddress) {
			logArgs = append(logArgs, "address", userAddress.AddressAsBech32String())
		}

		if debugErr == nil {
			logArgs = append(logArgs, "result", "success")
		} else {
			if strings.Contains(debugErr.Error(), tokensMismatchError) {
				logArgs = append(logArgs, "code", request.Code)
			}
			logArgs = append(logArgs, "error", debugErr.Error())
		}

		log.Info("Request info", logArgs...)
	}()

	userAddress, err := gg.extractAddressContext(c)
	if err != nil {
		debugErr = fmt.Errorf("%w while extracting user address", err)
		guardianLog.Debug("cannot extract user address for verify guardian", "error", err.Error())
		returnStatus(c, nil, http.StatusBadRequest, err.Error(), chainApiShared.ReturnCodeRequestError)
		return
	}

	err = json.NewDecoder(c.Request.Body).Decode(&request)
	if err != nil {
		debugErr = fmt.Errorf("%w while decoding request", err)
		guardianLog.Debug("cannot decode verify guardian request",
			"userAddress", userAddress.AddressAsBech32String(),
			"error", err.Error())
		returnStatus(c, nil, http.StatusBadRequest, err.Error(), chainApiShared.ReturnCodeRequestError)
		return
	}

	err = gg.facade.VerifyCode(userAddress, userIp, request)
	if err != nil {
		debugErr = fmt.Errorf("%w while verifying code", err)
		guardianLog.Debug("cannot verify guardian",
			"userAddress", userAddress.AddressAsBech32String(),
			"guardian", request.Guardian,
			"error", err.Error())
		returnStatus(c, nil, http.StatusInternalServerError, err.Error(), chainApiShared.ReturnCodeInternalError)
		return
	}

	returnStatus(c, nil, http.StatusOK, "", chainApiShared.ReturnCodeSuccess)
}

func (gg *guardianGroup) registeredUsers(c *gin.Context) {
	retData := &requests.RegisteredUsersResponse{}
	var err error

	userIp := c.GetString(mfaMiddleware.UserIpKey)
	userAgent := c.GetString(mfaMiddleware.UserAgentKey)
	defer func() {
		logArgs := []interface{}{
			"route", registeredUsersPath,
			"ip", userIp,
			"user agent", userAgent,
		}

		if err == nil {
			logArgs = append(logArgs, "result", "success")
			logArgs = append(logArgs, "returned count", retData.Count)
		} else {
			logArgs = append(logArgs, "error", err.Error())
		}

		log.Info("Request info", logArgs...)
	}()

	retData.Count, err = gg.facade.RegisteredUsers()
	if err != nil {
		guardianLog.Debug("cannot get number of registered users", "error", err.Error())
		returnStatus(c, nil, http.StatusInternalServerError, err.Error(), chainApiShared.ReturnCodeInternalError)
		return
	}

	returnStatus(c, retData, http.StatusOK, "", chainApiShared.ReturnCodeSuccess)
}

func (gg *guardianGroup) config(c *gin.Context) {
	userIp := c.GetString(mfaMiddleware.UserIpKey)
	userAgent := c.GetString(mfaMiddleware.UserAgentKey)
	log.Info("Request info", "route", tcsConfig, "ip", userIp, "user agent", userAgent)

	retData := &requests.ConfigResponse{}
	config := gg.facade.TcsConfig()
	retData.BackoffWrongCode = uint32(config.BackoffWrongCode)
	retData.RegistrationDelay = uint32(config.OTPDelay)

	returnStatus(c, retData, http.StatusOK, "", chainApiShared.ReturnCodeSuccess)
}

func returnStatus(c *gin.Context, data interface{}, httpStatus int, err string, code chainApiShared.ReturnCode) {
	c.JSON(
		httpStatus,
		chainApiShared.GenericAPIResponse{
			Data:  data,
			Error: err,
			Code:  code,
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

func (gg *guardianGroup) extractAddressContext(c *gin.Context) (core.AddressHandler, error) {
	userAddressStr := c.GetString(mfaMiddleware.UserAddressKey)
	return data.NewAddressFromBech32String(userAddressStr)
}

func getPrintableTxData(txData interface{}) string {
	txDataBuff, err := json.Marshal(txData)
	if err != nil {
		log.Warn("could not get printable txs", "error", err.Error())
		return ""
	}

	return string(txDataBuff)
}

// IsInterfaceNil returns true if there is no value under the interface
func (gg *guardianGroup) IsInterfaceNil() bool {
	return gg == nil
}
