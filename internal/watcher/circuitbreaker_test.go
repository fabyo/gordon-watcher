package watcher

import (
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker_Success(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)

	// Successful calls should pass through
	err := cb.Call(func() error {
		return nil
	})

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cb.GetState() != StateClosed {
		t.Errorf("Expected state Closed, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_OpensAfterFailures(t *testing.T) {
	cb := NewCircuitBreaker(3, 1*time.Second)

	testErr := errors.New("test error")

	// First 2 failures should keep circuit closed
	for i := 0; i < 2; i++ {
		err := cb.Call(func() error {
			return testErr
		})
		if err != testErr {
			t.Errorf("Expected test error, got %v", err)
		}
		if cb.GetState() != StateClosed {
			t.Errorf("Expected state Closed after %d failures, got %v", i+1, cb.GetState())
		}
	}

	// 3rd failure should open the circuit
	err := cb.Call(func() error {
		return testErr
	})
	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}
	if cb.GetState() != StateOpen {
		t.Errorf("Expected state Open after 3 failures, got %v", cb.GetState())
	}

	// Further calls should fail immediately without calling the function
	called := false
	err = cb.Call(func() error {
		called = true
		return nil
	})
	if err == nil {
		t.Error("Expected error when circuit is open")
	}
	if called {
		t.Error("Function should not be called when circuit is open")
	}
}

func TestCircuitBreaker_ClosesAfterTimeout(t *testing.T) {
	cb := NewCircuitBreaker(2, 100*time.Millisecond)

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return testErr
		})
	}

	if cb.GetState() != StateOpen {
		t.Fatalf("Expected state Open, got %v", cb.GetState())
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	// Next call should transition to HalfOpen
	err := cb.Call(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error after timeout, got %v", err)
	}

	// After successful call in HalfOpen, should be Closed
	if cb.GetState() != StateClosed {
		t.Errorf("Expected state Closed after successful recovery, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_HalfOpenRecovery(t *testing.T) {
	cb := NewCircuitBreaker(2, 100*time.Millisecond)

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return testErr
		})
	}

	// Wait for timeout to enter HalfOpen
	time.Sleep(150 * time.Millisecond)

	// Successful call should close the circuit
	err := cb.Call(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if cb.GetState() != StateClosed {
		t.Errorf("Expected state Closed, got %v", cb.GetState())
	}

	// Circuit should now accept calls normally
	err = cb.Call(func() error {
		return nil
	})
	if err != nil {
		t.Errorf("Expected no error after recovery, got %v", err)
	}
}

func TestCircuitBreaker_HalfOpenFailure(t *testing.T) {
	cb := NewCircuitBreaker(2, 100*time.Millisecond)

	testErr := errors.New("test error")

	// Open the circuit
	for i := 0; i < 2; i++ {
		cb.Call(func() error {
			return testErr
		})
	}

	// Wait for timeout to enter HalfOpen
	time.Sleep(150 * time.Millisecond)

	// Failed call in HalfOpen should keep circuit open
	err := cb.Call(func() error {
		return testErr
	})
	if err != testErr {
		t.Errorf("Expected test error, got %v", err)
	}

	// Circuit should still be open (not enough failures to re-open, but not closed either)
	// The implementation opens on >= maxFailures, so after 1 failure it's still HalfOpen
	// After 2 failures it should be Open again
	cb.Call(func() error {
		return testErr
	})

	if cb.GetState() != StateOpen {
		t.Errorf("Expected state Open after failures in HalfOpen, got %v", cb.GetState())
	}
}

func TestCircuitBreaker_Concurrent(t *testing.T) {
	cb := NewCircuitBreaker(5, 100*time.Millisecond)

	// Test concurrent access doesn't cause race conditions
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				cb.Call(func() error {
					if j%2 == 0 {
						return nil
					}
					return errors.New("test error")
				})
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Just verify no panic occurred
	_ = cb.GetState()
}
