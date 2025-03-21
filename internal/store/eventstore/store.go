package eventstore

import (
	"slices"
	"sync"

	"github.com/ubccr/grendel/pkg/model"
)

type Store struct {
	model.EventList
	Mu sync.RWMutex
}

func (s *Store) GetEvents() model.EventList {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	events := s.EventList
	slices.Reverse(events)
	return events
}

func (s *Store) StoreEvents(newEvent model.Event) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.EventList = append(s.EventList, newEvent)

	if len(s.EventList) > 50 {
		s.EventList = s.EventList[:50]
	}
}
