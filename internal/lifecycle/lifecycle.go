package lifecycle

import (
	"sync"
	"sync/atomic"
)

// State represents the application lifecycle state
type State int32

const (
	// StateStarting indicates the application is starting up
	StateStarting State = iota
	// StateReady indicates the application is ready to serve requests
	StateReady
	// StateShuttingDown indicates the application is shutting down
	StateShuttingDown
	// StateShutdown indicates the application has completed shutdown
	StateShutdown
)

// Manager manages the application lifecycle state
type Manager struct {
	state int32 // atomic access
	mu    sync.RWMutex
}

// NewManager creates a new lifecycle manager
func NewManager() *Manager {
	return &Manager{
		state: int32(StateStarting),
	}
}

// SetState sets the lifecycle state
func (m *Manager) SetState(s State) {
	atomic.StoreInt32(&m.state, int32(s))
}

// GetState returns the current lifecycle state
func (m *Manager) GetState() State {
	return State(atomic.LoadInt32(&m.state))
}

// IsReady returns true if the application is ready to serve requests
func (m *Manager) IsReady() bool {
	return m.GetState() == StateReady
}

// IsShuttingDown returns true if the application is shutting down
func (m *Manager) IsShuttingDown() bool {
	return m.GetState() == StateShuttingDown || m.GetState() == StateShutdown
}

// String returns a string representation of the state
func (s State) String() string {
	switch s {
	case StateStarting:
		return "starting"
	case StateReady:
		return "ready"
	case StateShuttingDown:
		return "shutting_down"
	case StateShutdown:
		return "shutdown"
	default:
		return "unknown"
	}
}

