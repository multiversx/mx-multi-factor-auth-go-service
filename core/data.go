package core

// GuardianState represents the state of the guardian
type GuardianState uint32

const (
	// NotUsableYet represents a guardian recently generated but not verified
	NotUsableYet GuardianState = iota
	// Usable represents a verified guardian
	Usable
)

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

// GuardianInfo holds details about a guardian
type GuardianInfo struct {
	PublicKey  []byte
	PrivateKey []byte
	State      GuardianState
}

// UserInfo holds info about both user's guardians and its unique index
type UserInfo struct {
	Index          uint32
	FirstGuardian  GuardianInfo
	SecondGuardian GuardianInfo
	Provider       string
}
