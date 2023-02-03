//go:generate protoc -I=. -I=$GOPATH/src -I=$GOPATH/src/github.com/multiversx/protobuf/protobuf  --gogoslick_out=. userInfo.proto

package core

// OnChainGuardianState represents the on chain state of the guardian
type OnChainGuardianState uint32

const (
	// ActiveGuardian represents an active guardian on chain
	ActiveGuardian OnChainGuardianState = iota
	// PendingGuardian represents a pending guardian on chain
	PendingGuardian
	// MissingGuardian represents a guardian missing from chain
	MissingGuardian
)
