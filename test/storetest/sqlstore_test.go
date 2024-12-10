// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package storetest

import (
	"path"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/sqlstore"
)

type SqlStoreTestSuite struct {
	StoreTestSuite
	file string
}

func TestSqlStoreTestSuite(t *testing.T) {
	store.Log.Logger.SetLevel(logrus.ErrorLevel)
	suite.Run(t, new(SqlStoreTestSuite))
}

func (s *SqlStoreTestSuite) SetFile(file string) {
	if file != ":memory:" {
		s.file = "file:" + file
	}
}

func (s *SqlStoreTestSuite) SetupTest() {
	file := ":memory:"
	if s.file != "" {
		file = s.file
	}
	var err error
	ds, err := sqlstore.New(file)
	s.Assert().NoError(err)
	s.SetStore(ds)
}

func TestSqlStoreMigrations(t *testing.T) {
	assert := assert.New(t)

	file := path.Join(t.TempDir(), "grendel-test.db")

	var err error
	_, err = sqlstore.New(file)
	assert.NoError(err)
	_, err = sqlstore.New(file)
	assert.NoError(err)
	_, err = sqlstore.New(file)
	assert.NoError(err)
}
