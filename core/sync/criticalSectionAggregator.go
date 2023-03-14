package sync

import "sync"

type criticalSectionsAggregator struct {
	mut              sync.RWMutex
	criticalSections map[string]*criticalSection
}

func (csa *criticalSectionsAggregator) Lock(key string) {
	csa.getSection(key).Lock()
}

func (csa *criticalSectionsAggregator) Unlock(key string) {
	section := csa.getSection(key)
	section.Unlock()

	csa.mut.Lock()
	defer csa.mut.Unlock()
	if section.NumLocks() == 0 {
		delete(csa.criticalSections, key)
	}
}

// NewCriticalSectionsAggregator returns a new instance of CriticalSectionsAggregator
func NewCriticalSectionsAggregator() CriticalSectionsAggregator {
	return &criticalSectionsAggregator{
		criticalSections: make(map[string]*criticalSection),
	}
}

// getSection returns the critical criticalSection for the given key
func (csa *criticalSectionsAggregator) getSection(key string) CriticalSection {
	csa.mut.RLock()
	section, ok := csa.criticalSections[key]
	csa.mut.RUnlock()
	if ok {
		return section
	}

	csa.mut.Lock()
	section, ok = csa.criticalSections[key]
	if !ok {
		section = &criticalSection{}
		csa.criticalSections[key] = section
	}
	csa.mut.Unlock()

	return section
}

// IsInterfaceNil returns true if there is no value under the interface
func (csa *criticalSectionsAggregator) IsInterfaceNil() bool {
	return csa == nil
}
