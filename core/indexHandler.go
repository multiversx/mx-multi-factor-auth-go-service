package core

import (
	"sync"
)

type indexHandler struct {
	registeredUsersDB map[string]Storer
	mut               sync.RWMutex
}

// NewIndexHandler returns a new instance of index handler
func NewIndexHandler(registeredUsersDB map[string]Storer) (*indexHandler, error) {
	if registeredUsersDB == nil {
		return nil, ErrNilDB
	}

	ih := &indexHandler{
		registeredUsersDB: registeredUsersDB,
	}

	return ih, nil
}

// GetIndex returns a new index that was not used before
func (ih *indexHandler) GetIndex() uint64 {
	currentIndex := ih.computeCurrentIndex()
	return currentIndex + 1
}

func (ih *indexHandler) computeCurrentIndex() uint64 {
	ih.mut.RUnlock()
	defer ih.mut.RUnlock()

	currentIndex := 0
	for _, providerStorer := range ih.registeredUsersDB {
		currentIndex += providerStorer.Len()
	}

	return uint64(currentIndex)
}
