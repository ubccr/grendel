// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/ubccr/grendel/internal/settings"
	"github.com/ubccr/grendel/internal/store/sqlstore/db"
)

func (s *SqlStore) Settings() ([]settings.Row, int64, error) {
	ctx := context.Background()
	// read in one transaction so the rows and the revision describe the same moment
	tx, err := s.ro.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, 0, err
	}
	defer tx.Rollback()

	rows, err := s.q.SettingList(ctx, tx)
	if err != nil {
		return nil, 0, err
	}
	revision, err := s.q.SettingRevision(ctx, tx)
	if err != nil {
		return nil, 0, err
	}

	out := make([]settings.Row, 0, len(rows))
	for _, row := range rows {
		out = append(out, settings.Row{
			Key:       row.Key,
			Value:     json.RawMessage(row.Value),
			UpdatedAt: row.UpdatedAt,
			UpdatedBy: row.UpdatedBy,
		})
	}
	return out, revision, nil
}

func (s *SqlStore) SettingsRevision() (int64, error) {
	return s.q.SettingRevision(context.Background(), s.ro)
}

func (s *SqlStore) InsertSettings(set []settings.Row, actor string) (int64, error) {
	var revision int64
	err := s.write(func(tx *sql.Tx) (err error) {
		ctx := context.Background()
		for _, row := range set {
			if err := s.q.SettingInsert(ctx, tx, db.SettingInsertParams{Key: row.Key, Value: string(row.Value), UpdatedBy: actor}); err != nil {
				return err
			}
		}
		revision, err = s.q.SettingRevisionBump(ctx, tx)
		return err
	})
	return revision, err
}

func (s *SqlStore) UpdateSettings(set []settings.Row, unset []string, actor string) (int64, error) {
	var revision int64
	err := s.write(func(tx *sql.Tx) (err error) {
		ctx := context.Background()
		for _, row := range set {
			if err := s.q.SettingUpsert(ctx, tx, db.SettingUpsertParams{Key: row.Key, Value: string(row.Value), UpdatedBy: actor}); err != nil {
				return err
			}
		}
		for _, key := range unset {
			if err := s.q.SettingDelete(ctx, tx, key); err != nil {
				return err
			}
		}
		revision, err = s.q.SettingRevisionBump(ctx, tx)
		return err
	})
	return revision, err
}
