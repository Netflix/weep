package health

import "sync"

var WeepStatus status

type status struct {
	healthy bool
	reason  string
	m       *sync.RWMutex
}

func init() {
	WeepStatus = status{
		healthy: true,
		reason:  "healthy",
		m:       &sync.RWMutex{},
	}
}

func (s *status) Get() (bool, string) {
	s.m.RLock()
	defer s.m.RUnlock()
	return s.healthy, s.reason
}

func (s *status) SetUnhealthy(reason string) {
	s.m.Lock()
	defer s.m.Unlock()
	s.healthy = false
	s.reason = reason
}

func (s *status) SetHealthy() {
	s.m.Lock()
	defer s.m.Unlock()
	s.healthy = true
	s.reason = "healthy"
}
