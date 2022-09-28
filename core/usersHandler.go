package core

import "sync"

type usersHandler struct {
	usersMut sync.RWMutex
	users    map[string]struct{}
}

// NewUsersHandler returns a new instance of usersHandler
func NewUsersHandler() *usersHandler {
	return &usersHandler{
		users: make(map[string]struct{}),
	}
}

// AddUser adds the provided address into the internal map
func (handler *usersHandler) AddUser(address string) {
	handler.usersMut.Lock()
	handler.users[address] = struct{}{}
	handler.usersMut.Unlock()
}

// HasUser returns true if the provided address is known
func (handler *usersHandler) HasUser(address string) bool {
	handler.usersMut.RLock()
	_, found := handler.users[address]
	handler.usersMut.RUnlock()

	return found
}

// RemoveUser removes the provided address from the internal map
func (handler *usersHandler) RemoveUser(address string) {
	if handler.HasUser(address) {
		handler.usersMut.Lock()
		delete(handler.users, address)
		handler.usersMut.Unlock()
	}
}

// IsInterfaceNil returns true if there is no value under the interface
func (handler *usersHandler) IsInterfaceNil() bool {
	return handler == nil
}
