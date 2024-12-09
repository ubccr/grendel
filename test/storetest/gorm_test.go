// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package storetest

import (
	"testing"

	"github.com/stretchr/testify/suite"
	gormstore "github.com/ubccr/grendel/internal/store/gorm"
)

type GormStoreTestSuite struct {
	StoreTestSuite
	file string
}

func TestGormStoreTestSuite(t *testing.T) {
	// store.Log.Logger.SetLevel(logrus.ErrorLevel)
	suite.Run(t, new(GormStoreTestSuite))
}

func (s *GormStoreTestSuite) SetFile(file string) {
	s.file = file
}

func (s *GormStoreTestSuite) SetupTest() {
	file := ":memory:"
	if s.file != "" {
		file = s.file
	}
	var err error
	ds, err := gormstore.New(file)
	s.Assert().NoError(err)
	s.SetStore(ds)
}

// func TestGormStoreMigrations(t *testing.T) {
// 	assert := assert.New(t)

// 	file := path.Join(t.TempDir(), "grendel-test-gorm.db")

// 	var err error
// 	_, err = gormstore.New(file)
// 	assert.NoError(err)
// 	_, err = gormstore.New(file)
// 	assert.NoError(err)
// 	_, err = gormstore.New(file)
// 	assert.NoError(err)
// }
