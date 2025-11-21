// internal/watcher/circuitbreaker.go (NOVO ARQUIVO!)
package watcher

import (
	"errors"
	"sync"
	"time"
)

type State int

const (
	StateClosed   State = iota // Normal
	StateOpen                  // Cortado (muitos erros)
	StateHalfOpen              // Testando recuperação
)

type CircuitBreaker struct {
	mu sync.Mutex

	maxFailures  int
	resetTimeout time.Duration

	failures    int
	lastFailure time.Time
	state       State
}

func NewCircuitBreaker(maxFailures int, resetTimeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		maxFailures:  maxFailures,
		resetTimeout: resetTimeout,
		state:        StateClosed,
	}
}

func (cb *CircuitBreaker) Call(fn func() error) error {
	cb.mu.Lock()

	// Se circuito aberto, verifica se pode testar
	if cb.state == StateOpen {
		if time.Since(cb.lastFailure) > cb.resetTimeout {
			cb.state = StateHalfOpen
			cb.failures = 0
		} else {
			cb.mu.Unlock()
			return errors.New("circuit breaker is open")
		}
	}

	cb.mu.Unlock()

	// Tenta executar
	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failures++
		cb.lastFailure = time.Now()

		if cb.failures >= cb.maxFailures {
			cb.state = StateOpen
		}

		return err
	}

	// Sucesso! Reseta
	if cb.state == StateHalfOpen {
		cb.state = StateClosed
	}
	cb.failures = 0

	return nil
}

func (cb *CircuitBreaker) GetState() State {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}
