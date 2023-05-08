//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. userInfo.proto

package core

// OnChainGuardianState represents the on chain state of the guardian
type OnChainGuardianState uint32

// TcsConfig represents the TCS configuration to be exposed to the user
type TcsConfig struct {
	OTPDelay         uint64
	BackoffWrongCode uint64
}

const (
	// ActiveGuardian represents an active guardian on chain
	ActiveGuardian OnChainGuardianState = iota
	// PendingGuardian represents a pending guardian on chain
	PendingGuardian
	// MissingGuardian represents a guardian missing from chain
	MissingGuardian
)
