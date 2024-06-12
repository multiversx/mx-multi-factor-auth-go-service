package testscommon

import (
	"github.com/multiversx/mx-chain-core-go/data/transaction"
	crypto "github.com/multiversx/mx-chain-crypto-go"
)

// SignerStub -
type SignerStub struct {
	SignMessageCalled     func(msg []byte, privateKey crypto.PrivateKey) ([]byte, error)
	VerifyMessageCalled   func(msg []byte, publicKey crypto.PublicKey, sig []byte) error
	SignTransactionCalled func(tx *transaction.FrontendTransaction, privateKey crypto.PrivateKey) ([]byte, error)
	SignByteSliceCalled   func(msg []byte, privateKey crypto.PrivateKey) ([]byte, error)
	VerifyByteSliceCalled func(msg []byte, publicKey crypto.PublicKey, sig []byte) error
}

// SignMessage -
func (s *SignerStub) SignMessage(msg []byte, privateKey crypto.PrivateKey) ([]byte, error) {
	if s.SignMessageCalled != nil {
		return s.SignMessageCalled(msg, privateKey)
	}

	return nil, nil
}

// VerifyMessage -
func (s *SignerStub) VerifyMessage(msg []byte, publicKey crypto.PublicKey, sig []byte) error {
	if s.SignMessageCalled != nil {
		return s.VerifyMessageCalled(msg, publicKey, sig)
	}

	return nil
}

// SignTransaction -
func (s *SignerStub) SignTransaction(tx *transaction.FrontendTransaction, privateKey crypto.PrivateKey) ([]byte, error) {
	if s.SignTransactionCalled != nil {
		return s.SignTransactionCalled(tx, privateKey)
	}

	return nil, nil
}

// SignByteSlice -
func (s *SignerStub) SignByteSlice(msg []byte, privateKey crypto.PrivateKey) ([]byte, error) {
	if s.SignByteSliceCalled != nil {
		return s.SignByteSliceCalled(msg, privateKey)
	}

	return nil, nil
}

// VerifyByteSlice -
func (s *SignerStub) VerifyByteSlice(msg []byte, publicKey crypto.PublicKey, sig []byte) error {
	if s.VerifyByteSliceCalled != nil {
		return s.VerifyByteSliceCalled(msg, publicKey, sig)
	}

	return nil
}

// IsInterfaceNil -
func (s *SignerStub) IsInterfaceNil() bool {
	return s == nil
}
