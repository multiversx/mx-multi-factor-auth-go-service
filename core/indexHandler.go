package core

import "github.com/ElrondNetwork/elrond-go-core/core/check"

type indexHandler struct {
	registeredUsersDB Storer
}

// NewIndexHandler returns a new instance of index handler
func NewIndexHandler(registeredUsersDB Storer) (*indexHandler, error) {
	if check.IfNil(registeredUsersDB) {
		return nil, ErrNilDB
	}

	ih := &indexHandler{
		registeredUsersDB: registeredUsersDB,
	}

	return ih, nil
}

// GetIndex returns a new index that was not used before
func (ih *indexHandler) GetIndex() uint32 {
	currentIndex := ih.registeredUsersDB.Len()
	return uint32(currentIndex) + 1
}

// IsInterfaceNil returns true if there is no value under the interface
func (ih *indexHandler) IsInterfaceNil() bool {
	return ih == nil
}
