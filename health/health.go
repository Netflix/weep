package health

import "sync"

var WeepStatus status

type status struct {
	sync.RWMutex
	healthy bool
	reason  string
}

func init() {
	WeepStatus = status{
		healthy: true,
		reason:  "healthy",
	}
}

func (s *status) Get() (bool, string) {
	s.RLock()
	defer s.RUnlock()
	return s.healthy, s.reason
}

func (s *status) SetUnhealthy(reason string) {
	s.Lock()
	defer s.Unlock()
	s.healthy = false
	s.reason = reason
}

func (s *status) SetHealthy() {
	s.Lock()
	defer s.Unlock()
	s.healthy = true
	s.reason = "healthy"
}
