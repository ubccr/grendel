// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package eventstore

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ubccr/grendel/pkg/model"
)

func store(s *Store, from, to int) {
	for i := from; i < to; i++ {
		s.StoreEvents(model.Event{Message: fmt.Sprintf("event-%d", i)})
	}
}

// assertNewestFirst checks the store holds exactly the events numbered from newest down to oldest
func assertNewestFirst(t *testing.T, s *Store, newest, oldest int) {
	t.Helper()

	events := s.GetEvents()
	if want := newest - oldest + 1; len(events) != want {
		t.Fatalf("got %d events, want %d", len(events), want)
	}
	for i, e := range events {
		if want := fmt.Sprintf("event-%d", newest-i); e.Message != want {
			t.Fatalf("position %d is %q, want %q", i, e.Message, want)
		}
	}
}

func TestKeepsTheLatestEvents(t *testing.T) {
	s := &Store{}

	store(s, 0, 60)

	assertNewestFirst(t, s, 59, 10)
}

func TestReadingDoesNotChangeTheOrder(t *testing.T) {
	s := &Store{}

	store(s, 0, 5)

	assertNewestFirst(t, s, 4, 0)
	assertNewestFirst(t, s, 4, 0)
}

// TestReadsBetweenWritesKeepTheLatest is the case a read that disturbed the stored order would break, since trimming would then drop the wrong end
func TestReadsBetweenWritesKeepTheLatest(t *testing.T) {
	s := &Store{}

	for i := 0; i < 60; i++ {
		store(s, i, i+1)
		s.GetEvents()
	}

	assertNewestFirst(t, s, 59, 10)
}

func TestReturnedEventsAreACopy(t *testing.T) {
	s := &Store{}

	store(s, 0, 3)

	s.GetEvents()[0].Message = "changed"

	assertNewestFirst(t, s, 2, 0)
}

func TestConcurrentUse(t *testing.T) {
	s := &Store{}

	var wg sync.WaitGroup
	for range 8 {
		wg.Add(2)
		go func() {
			defer wg.Done()
			store(s, 0, 100)
		}()
		go func() {
			defer wg.Done()
			for range 100 {
				s.GetEvents()
			}
		}()
	}
	wg.Wait()

	if got := len(s.GetEvents()); got != 50 {
		t.Errorf("got %d events, want 50", got)
	}
}
