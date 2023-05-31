package sync

// MutexType defines the mutex type
type MutexType string

const (
	// LocalMutex specifies local mutex
	LocalMutex MutexType = "local_mutex"

	// DistributedMutex specifies distributed mutex type
	DistributedMutex MutexType = "distributed_mutex"
)
