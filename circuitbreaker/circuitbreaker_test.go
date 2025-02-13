package circuitbreaker

import (
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	cb := NewCircuitBreaker(3, 2*time.Second)

	// Initial state should be Closed
	if !cb.Allow() {
		t.Error("Expected Allow to return true in Closed state")
	}

	// Simulate failures
	cb.Failure()
	cb.Failure()
	cb.Failure()

	// After 3 failures, state should be Open
	if cb.Allow() {
		t.Error("Expected Allow to return false in Open state")
	}

	// Wait for openTimeout to expire
	time.Sleep(3 * time.Second)

	// After timeout, state should be HalfOpen
	if !cb.Allow() {
		t.Error("Expected Allow to return true in HalfOpen state")
	}

	// Simulate a success in HalfOpen state
	cb.Success()

	// After a success, state should be Closed
	if !cb.Allow() {
		t.Error("Expected Allow to return true in Closed state")
	}

	// Simulate more failures
	cb.Failure()
	cb.Failure()
	cb.Failure()

	// After 3 failures again, state should be Open
	if cb.Allow() {
		t.Error("Expected Allow to return false in Open state")
	}
}
