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
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/multi-factor-auth-go-service/resolver"
	"github.com/multiversx/mx-chain-core-go/core/check"
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	chainApiShared "github.com/multiversx/mx-chain-go/api/shared"
	logger "github.com/multiversx/mx-chain-logger-go"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/data"
)

const (
	signTransactionPath          = "/sign-transaction"
	signMultipleTransactionsPath = "/sign-multiple-transactions"
	registerPath                 = "/register"
	verifyCodePath               = "/verify-code"
	registeredUsersPath          = "/registered-users"
	tcsConfig                    = "/config"

	wrongCodeError = "wrong code"
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
		return nil, fmt.Errorf("%w for guardian group", core.ErrNilFacadeHandler)
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
		logSignTransaction(userIp, userAgent, &request, debugErr)
	}()

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	if err != nil {
		debugErr = fmt.Errorf("%w while decoding request", err)
		returnStatus(c, nil, http.StatusBadRequest, err.Error(), chainApiShared.ReturnCodeRequestError)
		return
	}

	var signTransactionResponse *requests.SignTransactionResponse
	marshalledTx, otpCodeVerifyData, err := gg.facade.SignTransaction(userIp, request)
	if err != nil {
		debugErr = fmt.Errorf("%w while signing transaction", err)
		handleErrorAndReturn(c, getVerifyCodeResponse(otpCodeVerifyData), err.Error())
		return
	}

	signTransactionResponse, err = createSignTransactionResponse(marshalledTx)
	if err != nil {
		debugErr = fmt.Errorf("%w while creating response", err)
		returnStatus(c, nil, http.StatusInternalServerError, err.Error(), chainApiShared.ReturnCodeInternalError)
		return
	}

	returnStatus(c, signTransactionResponse, http.StatusOK, "", chainApiShared.ReturnCodeSuccess)
}

func logSignTransaction(userIp string, userAgent string, request *requests.SignTransaction, debugErr error) {
	logArgs := []interface{}{
		"route", signTransactionPath,
		"ip", userIp,
		"user agent", userAgent,
		"transaction", getPrintableTxData(request.Tx),
	}
	defer func() {
		guardianLog.Info("Request info", logArgs...)
	}()

	if debugErr == nil {
		logArgs = append(logArgs, "result", "success")
		return
	}

	if strings.Contains(debugErr.Error(), wrongCodeError) {
		logArgs = append(logArgs, "code", request.Code)
	}
	logArgs = append(logArgs, "error", debugErr.Error())
}

// signMultipleTransactions returns the transactions signed by the guardian if the verification passed
func (gg *guardianGroup) signMultipleTransactions(c *gin.Context) {
	var request requests.SignMultipleTransactions
	var debugErr error

	userIp := c.GetString(mfaMiddleware.UserIpKey)
	userAgent := c.GetString(mfaMiddleware.UserAgentKey)
	defer func() {
		logSignMultipleTransactions(userIp, userAgent, &request, debugErr)
	}()

	err := json.NewDecoder(c.Request.Body).Decode(&request)
	if err != nil {
		debugErr = fmt.Errorf("%w while decoding request", err)
		returnStatus(c, nil, http.StatusBadRequest, err.Error(), chainApiShared.ReturnCodeRequestError)
		return
	}

	marshalledTxs, otpCodeVerifyData, err := gg.facade.SignMultipleTransactions(userIp, request)
	if err != nil {
		debugErr = fmt.Errorf("%w while signing transactions", err)
		handleErrorAndReturn(c, getVerifyCodeResponse(otpCodeVerifyData), err.Error())
		return
	}

	var signMultipleTransactionsResponse *requests.SignMultipleTransactionsResponse
	signMultipleTransactionsResponse, err = createSignMultipleTransactionsResponse(marshalledTxs)
	if err != nil {
		debugErr = fmt.Errorf("%w while creating response", err)
		returnStatus(c, nil, http.StatusInternalServerError, err.Error(), chainApiShared.ReturnCodeInternalError)
		return
	}

	returnStatus(c, signMultipleTransactionsResponse, http.StatusOK, "", chainApiShared.ReturnCodeSuccess)
}

func logSignMultipleTransactions(userIp string, userAgent string, request *requests.SignMultipleTransactions, debugErr error) {
	logArgs := []interface{}{
		"route", signMultipleTransactionsPath,
		"ip", userIp,
		"user agent", userAgent,
		"transactions", getPrintableTxData(&request.Txs),
	}
	defer func() {
		guardianLog.Info("Request info", logArgs...)
	}()

	if debugErr == nil {
		logArgs = append(logArgs, "result", "success")
		return
	}

	if strings.Contains(debugErr.Error(), wrongCodeError) {
		logArgs = append(logArgs, "code", request.Code)
	}
	logArgs = append(logArgs, "error", debugErr.Error())
}

func createSignMultipleTransactionsResponse(marshalledTxs [][]byte) (*requests.SignMultipleTransactionsResponse, error) {
	signMultipleTransactionsResponse := &requests.SignMultipleTransactionsResponse{
		Txs: make([]transaction.FrontendTransaction, 0),
	}
	for i, marshalledTx := range marshalledTxs {
		unmarshalledTx := transaction.FrontendTransaction{}
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
	var userAddress sdkCore.AddressHandler
	retData := &requests.RegisterReturnData{}
	var debugErr error

	userIp := c.GetString(mfaMiddleware.UserIpKey)
	userAgent := c.GetString(mfaMiddleware.UserAgentKey)
	defer func() {
		logRegister(userIp, userAgent, userAddress, retData, debugErr)
	}()

	userAddress, err := gg.extractAddressContext(c)
	if err != nil {
		debugErr = fmt.Errorf("%w while extracting user address", err)
		returnStatus(c, nil, http.StatusBadRequest, err.Error(), chainApiShared.ReturnCodeRequestError)
		return
	}

	var request requests.RegistrationPayload
	err = json.NewDecoder(c.Request.Body).Decode(&request)
	if err != nil {
		debugErr = fmt.Errorf("%w while decoding request", err)
		returnStatus(c, nil, http.StatusBadRequest, err.Error(), chainApiShared.ReturnCodeRequestError)
		return
	}

	retData.OTP, retData.GuardianAddress, err = gg.facade.RegisterUser(userAddress, request)
	if err != nil {
		debugErr = fmt.Errorf("%w while registering", err)
		handleErrorAndReturn(c, retData, err.Error())
		return
	}

	returnStatus(c, retData, http.StatusOK, "", chainApiShared.ReturnCodeSuccess)
}

func getVerifyCodeResponse(verifyData *requests.OTPCodeVerifyData) requests.OTPCodeVerifyDataResponse {
	return requests.OTPCodeVerifyDataResponse{
		VerifyData: verifyData,
	}
}

func logRegister(userIp string, userAgent string, userAddress sdkCore.AddressHandler, retData *requests.RegisterReturnData, debugErr error) {
	logArgs := []interface{}{
		"route", registerPath,
		"ip", userIp,
		"user agent", userAgent,
	}
	defer func() {
		guardianLog.Info("Request info", logArgs...)
	}()

	if !check.IfNil(userAddress) {
		logArgs = append(logArgs, "address", userAddress.AddressAsBech32String())
	}

	if debugErr == nil {
		logArgs = append(logArgs, "result", "success")
		logArgs = append(logArgs, "returned guardian", retData.GuardianAddress)
		return
	}

	logArgs = append(logArgs, "error", debugErr.Error())
}

// verifyCode validates a code
func (gg *guardianGroup) verifyCode(c *gin.Context) {
	var request requests.VerificationPayload
	var userAddress sdkCore.AddressHandler
	var debugErr error

	userIp := c.GetString(mfaMiddleware.UserIpKey)
	userAgent := c.GetString(mfaMiddleware.UserAgentKey)
	defer func() {
		logVerifyCode(userIp, userAgent, userAddress, request, debugErr)
	}()

	userAddress, err := gg.extractAddressContext(c)
	if err != nil {
		debugErr = fmt.Errorf("%w while extracting user address", err)
		returnStatus(c, nil, http.StatusBadRequest, err.Error(), chainApiShared.ReturnCodeRequestError)
		return
	}

	err = json.NewDecoder(c.Request.Body).Decode(&request)
	if err != nil {
		debugErr = fmt.Errorf("%w while decoding request", err)
		returnStatus(c, nil, http.StatusBadRequest, err.Error(), chainApiShared.ReturnCodeRequestError)
		return
	}

	otpVerifyCodeData, err := gg.facade.VerifyCode(userAddress, userIp, request)
	if err != nil {
		debugErr = fmt.Errorf("%w while verifying code", err)
		handleErrorAndReturn(c, getVerifyCodeResponse(otpVerifyCodeData), err.Error())
		return
	}

	returnStatus(c, nil, http.StatusOK, "", chainApiShared.ReturnCodeSuccess)
}

func logVerifyCode(userIp string, userAgent string, userAddress sdkCore.AddressHandler, request requests.VerificationPayload, debugErr error) {
	logArgs := []interface{}{
		"route", verifyCodePath,
		"ip", userIp,
		"user agent", userAgent,
		"guardian", request.Guardian,
	}
	defer func() {
		guardianLog.Info("Request info", logArgs...)
	}()

	if !check.IfNil(userAddress) {
		logArgs = append(logArgs, "address", userAddress.AddressAsBech32String())
	}

	if debugErr == nil {
		logArgs = append(logArgs, "result", "success")
		return
	}

	if strings.Contains(debugErr.Error(), wrongCodeError) {
		logArgs = append(logArgs, "code", request.Code)
	}
	logArgs = append(logArgs, "error", debugErr.Error())
}

func (gg *guardianGroup) registeredUsers(c *gin.Context) {
	retData := &requests.RegisteredUsersResponse{}
	var err error
	retData.Count, err = gg.facade.RegisteredUsers()
	if err != nil {
		handleErrorAndReturn(c, nil, err.Error())
		return
	}

	returnStatus(c, retData, http.StatusOK, "", chainApiShared.ReturnCodeSuccess)
}

func (gg *guardianGroup) config(c *gin.Context) {
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

func handleErrorAndReturn(c *gin.Context, data interface{}, err string) {
	if strings.Contains(err, wrongCodeError) ||
		strings.Contains(err, resolver.ErrTooManyTransactionsToSign.Error()) ||
		strings.Contains(err, resolver.ErrNoTransactionToSign.Error()) ||
		strings.Contains(err, resolver.ErrGuardianMismatch.Error()) ||
		strings.Contains(err, resolver.ErrInvalidSender.Error()) ||
		strings.Contains(err, resolver.ErrInvalidGuardian.Error()) ||
		strings.Contains(err, resolver.ErrGuardianNotUsable.Error()) ||
		strings.Contains(err, resolver.ErrGuardianMismatch.Error()) {

		returnStatus(c, data, http.StatusBadRequest, err, chainApiShared.ReturnCodeRequestError)
		return
	}

	if strings.Contains(err, core.ErrTooManyFailedAttempts.Error()) {
		returnStatus(c, data, http.StatusTooManyRequests, err, chainApiShared.ReturnCodeRequestError)
		return
	}

	returnStatus(c, data, http.StatusInternalServerError, err, chainApiShared.ReturnCodeInternalError)
}

// UpdateFacade will update the facade
func (gg *guardianGroup) UpdateFacade(newFacade shared.FacadeHandler) error {
	if check.IfNil(newFacade) {
		return core.ErrNilFacadeHandler
	}

	gg.mutFacade.Lock()
	gg.facade = newFacade
	gg.mutFacade.Unlock()

	return nil
}

func (gg *guardianGroup) extractAddressContext(c *gin.Context) (sdkCore.AddressHandler, error) {
	userAddressStr := c.GetString(mfaMiddleware.UserAddressKey)
	return data.NewAddressFromBech32String(userAddressStr)
}

func getPrintableTxData(txData interface{}) string {
	txDataBuff, err := json.Marshal(txData)
	if err != nil {
		guardianLog.Warn("could not get printable txs", "error", err.Error())
		return ""
	}

	return string(txDataBuff)
}

// IsInterfaceNil returns true if there is no value under the interface
func (gg *guardianGroup) IsInterfaceNil() bool {
	return gg == nil
}
