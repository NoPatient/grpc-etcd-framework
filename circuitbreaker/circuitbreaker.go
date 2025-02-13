package circuitbreaker

import (
	"sync"
	"time"
)

type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

type CircuitBreaker struct {
	mu               sync.Mutex
	failures         int
	state            State
	failureThreshold int
	openTimeout      time.Duration
	lastFailureTime  time.Time
}

func NewCircuitBreaker(failureThreshold int, openTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:            Closed,
		failureThreshold: failureThreshold,
		openTimeout:      openTimeout,
	}
}

func (c *CircuitBreaker) Allow() bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.state {
	case Closed:
		return true
	case Open:
		if time.Since(c.lastFailureTime) > c.openTimeout {
			c.state = HalfOpen
			return true
		}
		return false
	case HalfOpen:
		return true
	default:
		return false
	}
}

func (c *CircuitBreaker) Success() {
	c.mu.Lock()
	c.mu.Unlock()

	c.failures = 0
	c.state = Closed
}

func (c *CircuitBreaker) Failure() {
	c.mu.Lock()
	c.mu.Unlock()

	c.failures++
	if c.failures >= c.failureThreshold {
		c.state = Open
		c.lastFailureTime = time.Now()
	}
}
