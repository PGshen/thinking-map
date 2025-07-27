package react

import "sync"

// UnboundedChan represents a channel with unlimited capacity
type UnboundedChan[T any] struct {
	buffer   []T        // Internal buffer to store data
	mutex    sync.Mutex // Mutex to protect buffer access
	notEmpty *sync.Cond // Condition variable to wait for data
	closed   bool       // Indicates if the channel has been closed
}

// NewUnboundedChan initializes and returns an UnboundedChan
func NewUnboundedChan[T any]() *UnboundedChan[T] {
	ch := &UnboundedChan[T]{}
	ch.notEmpty = sync.NewCond(&ch.mutex)
	return ch
}

// Send puts an item into the channel
func (ch *UnboundedChan[T]) Send(value T) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	if ch.closed {
		panic("send on closed channel")
	}

	ch.buffer = append(ch.buffer, value)
	ch.notEmpty.Signal() // Wake up one goroutine waiting to receive
}

// Receive gets an item from the channel (blocks if empty)
func (ch *UnboundedChan[T]) Receive() (T, bool) {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	for len(ch.buffer) == 0 && !ch.closed {
		ch.notEmpty.Wait() // Wait until data is available
	}

	if len(ch.buffer) == 0 {
		// Channel is closed and empty
		var zero T
		return zero, false
	}

	val := ch.buffer[0]
	ch.buffer = ch.buffer[1:]
	return val, true
}

// Close marks the channel as closed
func (ch *UnboundedChan[T]) Close() {
	ch.mutex.Lock()
	defer ch.mutex.Unlock()

	if !ch.closed {
		ch.closed = true
		ch.notEmpty.Broadcast() // Wake up all waiting goroutines
	}
}
