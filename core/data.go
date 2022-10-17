package core

// GuardianState represents the state of the guardian
type GuardianState uint32

// Usable represents a verified guardian
const Usable GuardianState = 1

// NotUsableYet represents a guardian recently generated but not verified
const NotUsableYet GuardianState = 0

// OnChainGuardianState represents the on chain state of the guardian
type OnChainGuardianState uint32

// ActiveGuardian represents an active guardian on chain
const ActiveGuardian OnChainGuardianState = 0

// PendingGuardian represents a pending guardian on chain
const PendingGuardian OnChainGuardianState = 1

// MissingGuardian represents a guardian missing from chain
const MissingGuardian OnChainGuardianState = 2

// GuardianInfo holds details about a guardian
type GuardianInfo struct {
	PublicKey  []byte
	PrivateKey []byte
	State      GuardianState
}

// UserInfo holds info about both user's guardians and its unique index
type UserInfo struct {
	Index            uint32
	FirstGuardian    GuardianInfo
	SecondarGuardian GuardianInfo
	Provider         string
}
