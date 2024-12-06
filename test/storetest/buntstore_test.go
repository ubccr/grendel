// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package storetest

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/ubccr/grendel/internal/store/buntstore"
)

type BuntStoreTestSuite struct {
	StoreTestSuite
	file string
}

func TestBuntStoreTestSuite(t *testing.T) {
	suite.Run(t, new(BuntStoreTestSuite))
}

func (s *BuntStoreTestSuite) SetFile(file string) {
	s.file = file
}

func (s *BuntStoreTestSuite) SetupTest() {
	file := ":memory:"
	if s.file != "" {
		file = s.file
	}
	var err error
	ds, err := buntstore.New(file)
	s.Assert().NoError(err)
	s.SetStore(ds)
}
