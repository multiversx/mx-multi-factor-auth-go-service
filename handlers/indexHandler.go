package handlers

import (
	"sync"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
)

type indexHandler struct {
	registeredUsersDB core.Storer
	marshaller        core.Marshaller
	latestIndex       uint32
	indexMut          sync.Mutex
}

// NewIndexHandler returns a new instance of index handler
func NewIndexHandler(registeredUsersDB core.Storer, marshaller core.Marshaller) (*indexHandler, error) {
	if check.IfNil(registeredUsersDB) {
		return nil, ErrNilDB
	}
	if check.IfNil(marshaller) {
		return nil, ErrNilMarshaller
	}

	ih := &indexHandler{
		registeredUsersDB: registeredUsersDB,
		marshaller:        marshaller,
	}

	err := ih.fetchLatestIndex()
	if err != nil {
		return nil, err
	}

	return ih, err
}

func (ih *indexHandler) fetchLatestIndex() error {
	var err error
	ih.registeredUsersDB.RangeKeys(func(key []byte, val []byte) bool {
		userInfo := &core.UserInfo{}
		err = ih.marshaller.Unmarshal(userInfo, val)
		if err != nil {
			return false
		}

		if ih.latestIndex < userInfo.Index {
			ih.latestIndex = userInfo.Index
		}

		return true
	})
	return err
}

// AllocateIndex returns a new index that was not used before
func (ih *indexHandler) AllocateIndex() uint32 {
	ih.indexMut.Lock()
	defer ih.indexMut.Unlock()
	ih.latestIndex++
	return ih.latestIndex
}

// RevertIndex reverts the index to previous value
func (ih *indexHandler) RevertIndex() {
	ih.indexMut.Lock()
	defer ih.indexMut.Unlock()
	if ih.latestIndex > 0 {
		ih.latestIndex--
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (ih *indexHandler) IsInterfaceNil() bool {
	return ih == nil
}
