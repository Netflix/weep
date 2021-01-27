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

func (s *status) Set(healthy bool, reason string) {
	s.m.Lock()
	defer s.m.Unlock()
	s.healthy = healthy
	s.reason = reason
}
