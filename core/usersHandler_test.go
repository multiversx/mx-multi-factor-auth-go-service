package core

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/ElrondNetwork/elrond-go-core/core/check"
	"github.com/stretchr/testify/assert"
)

func TestUsersHandler(t *testing.T) {
	t.Parallel()

	handler := NewUsersHandler()
	assert.False(t, check.IfNil(handler))

	providedUser := "provided user"
	_ = handler.AddUser(providedUser)
	assert.True(t, handler.HasUser(providedUser))
	handler.RemoveUser(providedUser)
	assert.False(t, handler.HasUser(providedUser))
}

func TestUsersHandler_Concurrency(t *testing.T) {
	t.Parallel()

	defer func() {
		r := recover()
		if r != nil {
			assert.Fail(t, "should not panic")
		}
	}()

	handler := NewUsersHandler()
	assert.False(t, check.IfNil(handler))

	numOperations := 1000
	var wg sync.WaitGroup
	wg.Add(numOperations)

	for idx := 0; idx < numOperations; idx++ {
		go func(index int) {
			time.Sleep(time.Millisecond)
			address := fmt.Sprintf("address_%d", index)
			switch index % 3 {
			case 0:
				_ = handler.AddUser(address)
			case 1:
				handler.HasUser(address)
			case 2:
				handler.RemoveUser(address)
			}
			wg.Done()
		}(idx)
	}

	wg.Wait()
}
