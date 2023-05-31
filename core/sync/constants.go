package sync

// LockingType defines the locking type
type LockingType string

const (
	// LocalMutex specifies local mutex
	LocalMutex LockingType = "local_mutex"

	// DistributedMutex specifies distributed mutex type
	DistributedMutex LockingType = "distributed_mutex"
)
