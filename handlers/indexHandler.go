package handlers

import (
	"encoding/binary"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/ElrondNetwork/multi-factor-auth-go-service/core"
)

const (
	lastIndexKey = "lastAllocatedIndex"
	uint32Bytes  = 4
)

type indexHandler struct {
	registeredUsersDB core.Storer
}

// NewIndexHandler returns a new instance of index handler
func NewIndexHandler(registeredUsersDB core.Storer) (*indexHandler, error) {
	if check.IfNil(registeredUsersDB) {
		return nil, ErrNilDB
	}

	ih := &indexHandler{
		registeredUsersDB: registeredUsersDB,
	}
	err := ih.registeredUsersDB.Has([]byte(lastIndexKey))
	if err != nil {
		err = ih.saveNewIndex(0)
	}

	return ih, err
}

// AllocateIndex returns a new index that was not used before
func (ih *indexHandler) AllocateIndex() (uint32, error) {
	lastIndex, err := ih.getIndex()
	if err != nil {
		return 0, err
	}
	lastIndex++

	err = ih.saveNewIndex(lastIndex)
	if err != nil {
		return 0, err
	}

	return lastIndex, nil
}

func (ih *indexHandler) getIndex() (uint32, error) {
	lastIndexBytes, err := ih.registeredUsersDB.Get([]byte(lastIndexKey))
	if err != nil {
		return 0, err
	}

	return binary.BigEndian.Uint32(lastIndexBytes), nil
}

func (ih *indexHandler) saveNewIndex(newIndex uint32) error {
	latestIndexBytes := make([]byte, uint32Bytes)
	binary.BigEndian.PutUint32(latestIndexBytes, newIndex)
	return ih.registeredUsersDB.Put([]byte(lastIndexKey), latestIndexBytes)
}

// IsInterfaceNil returns true if there is no value under the interface
func (ih *indexHandler) IsInterfaceNil() bool {
	return ih == nil
}
