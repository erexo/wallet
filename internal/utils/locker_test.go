package utils

import (
	"testing"
	"time"
)

func TestLocker(t *testing.T) {
	locker := NewRWLocker[int]()

	// multiple rlocks
	unlock11 := locker.RLock(1)
	unlock12 := locker.RLock(1)

	// lock on different values
	unlock2 := locker.Lock(2)
	unlock3 := locker.Lock(3)

	unlock11()
	unlock12()
	unlock2()

	go func() {
		time.Sleep(100 * time.Millisecond)
		unlock3()
	}()

	locker.Wait()
}
