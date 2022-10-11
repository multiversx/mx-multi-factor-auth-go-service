package core

// GuardianState represents the state of the guardian
type GuardianState string

// Usable represents a verified guardian
const Usable GuardianState = "usable"

// NotUsableYet represents a guardian recently generated but not verified
const NotUsableYet GuardianState = "notUsableYet"

// OnChainGuardianState represents the on chain state of the guardian
type OnChainGuardianState string

// ActiveGuardian represents an active guardian on chain
const ActiveGuardian OnChainGuardianState = "activeGuardian"

// PendingGuardian represents a pending guardian on chain
const PendingGuardian OnChainGuardianState = "pendingGuardian"

// MissingGuardian represents a guardian missing from chain
const MissingGuardian OnChainGuardianState = "missingGuardian"

// GuardianInfo holds details about a guardian
type GuardianInfo struct {
	PublicKey  []byte
	PrivateKey []byte
	State      GuardianState
}

// UserInfo holds info about both user's guardians and its unique index
type UserInfo struct {
	Index             uint32
	MainGuardian      GuardianInfo
	SecondaryGuardian GuardianInfo
	Provider          string
}
