package facade

import (
	"github.com/multiversx/multi-factor-auth-go-service/core"
	"github.com/multiversx/multi-factor-auth-go-service/core/requests"
	"github.com/multiversx/mx-chain-core-go/core/check"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
)

// ArgsGuardianFacade represents the DTO struct used in the auth facade constructor
type ArgsGuardianFacade struct {
	ServiceResolver core.ServiceResolver
}

type guardianFacade struct {
	serviceResolver core.ServiceResolver
}

// NewGuardianFacade returns a new instance of guardianFacade
func NewGuardianFacade(args ArgsGuardianFacade) (*guardianFacade, error) {
	if check.IfNil(args.ServiceResolver) {
		return nil, ErrNilServiceResolver
	}

	return &guardianFacade{
		serviceResolver: args.ServiceResolver,
	}, nil
}

// VerifyCode validates the code received
func (af *guardianFacade) VerifyCode(userAddress sdkCore.AddressHandler, userIp string, request requests.VerificationPayload) error {
	return af.serviceResolver.VerifyCode(userAddress, userIp, request)
}

// RegisterUser creates a new OTP and (optionally) returns some information required
// for the user to set up the OTP on his end (eg: QR code).
func (af *guardianFacade) RegisterUser(userAddress sdkCore.AddressHandler, request requests.RegistrationPayload) ([]byte, string, error) {
	return af.serviceResolver.RegisterUser(userAddress, request)
}

// SignTransaction validates user's transaction, then signs it from guardian and returns the transaction
func (af *guardianFacade) SignTransaction(userAddress sdkCore.AddressHandler, userIp string, request requests.SignTransaction) ([]byte, error) {
	return af.serviceResolver.SignTransaction(userAddress, userIp, request)
}

// SignMultipleTransactions validates user's transactions, then adds guardian signature and returns the transaction
func (af *guardianFacade) SignMultipleTransactions(userAddress sdkCore.AddressHandler, userIp string, request requests.SignMultipleTransactions) ([][]byte, error) {
	return af.serviceResolver.SignMultipleTransactions(userAddress, userIp, request)
}

// RegisteredUsers returns the number of registered users
func (af *guardianFacade) RegisteredUsers() (uint32, error) {
	return af.serviceResolver.RegisteredUsers()
}

// IsInterfaceNil returns true if there is no value under the interface
func (af *guardianFacade) IsInterfaceNil() bool {
	return af == nil
}
