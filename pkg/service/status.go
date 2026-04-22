package service

import (
	"sync"

	"github.com/gauffer/ddd-events-codegen-converters/pkg/log"

	"go.uber.org/zap"
)

type Status struct {
	alive bool
	ready bool

	logger log.Logger

	mu sync.RWMutex
}

func (s *Status) SetAlive(state bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.alive = state
	s.logger.With(zap.Bool("alive", state)).Info("alive status changed")
}

func (s *Status) SetReady(state bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ready = state
	s.logger.With(zap.Bool("ready", state)).Info("ready status changed")
}

func (s *Status) IsReady() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ready
}

func (s *Status) IsAlive() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.alive
}
