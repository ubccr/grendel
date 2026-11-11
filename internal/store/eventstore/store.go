package eventstore

import (
	"slices"
	"sync"

	"github.com/ubccr/grendel/pkg/model"
)

// maxEvents is how many of the latest events are kept
const maxEvents = 50

type Store struct {
	model.EventList
	Mu sync.RWMutex
}

// GetEvents returns a copy of the events, newest first
func (s *Store) GetEvents() model.EventList {
	s.Mu.RLock()
	defer s.Mu.RUnlock()

	events := slices.Clone(s.EventList)
	slices.Reverse(events)
	return events
}

func (s *Store) StoreEvents(newEvent model.Event) {
	s.Mu.Lock()
	defer s.Mu.Unlock()

	s.EventList = append(s.EventList, newEvent)

	if len(s.EventList) > maxEvents {
		s.EventList = s.EventList[len(s.EventList)-maxEvents:]
	}
}
