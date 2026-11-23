// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package sqlstore

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/segmentio/ksuid"
	"github.com/ubccr/grendel/internal/store"
	"github.com/ubccr/grendel/internal/store/sqlstore/db"
	"github.com/ubccr/grendel/pkg/model"
	"github.com/ubccr/grendel/pkg/nodeset"
)

func (s *SqlStore) LoadProvisionStatus(uid string) (*model.ProvisionStatus, error) {
	id, err := ksuid.Parse(uid)
	if err != nil {
		return nil, err
	}

	row, err := s.q.ProvisionStatusFetch(context.Background(), s.ro, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, store.ErrNotFound
		}
		return nil, err
	}

	return provisionStatusFromRow(row)
}

// The status is read and written in one transaction on the write connection, so two reports for a host never start from the same stored status
func (s *SqlStore) UpdateProvisionStatus(uid string, open *model.Event, update func(status *model.ProvisionStatus) (*model.EventNode, error)) error {
	id, err := ksuid.Parse(uid)
	if err != nil {
		return err
	}

	return s.write(func(tx *sql.Tx) error {
		ctx := context.Background()

		status := &model.ProvisionStatus{}
		row, err := s.q.ProvisionStatusFetch(ctx, tx, id)
		switch {
		case err == nil:
			if status, err = provisionStatusFromRow(row); err != nil {
				return err
			}
		case !errors.Is(err, sql.ErrNoRows):
			return err
		}

		node, err := update(status)
		if err != nil {
			return err
		}

		detail, err := json.Marshal(status.ProvisionStatusDetail)
		if err != nil {
			return err
		}
		completed := sql.NullTime{}
		if status.CompletedAt != nil {
			completed = sql.NullTime{Time: status.CompletedAt.UTC(), Valid: true}
		}
		n, err := s.q.ProvisionStatusUpsert(ctx, tx, db.ProvisionStatusUpsertParams{
			UID:         id,
			State:       string(status.State),
			StartedAt:   status.StartedAt.UTC(),
			UpdatedAt:   status.UpdatedAt.UTC(),
			CompletedAt: completed,
			Detail:      string(detail),
		})
		if err != nil {
			return err
		}
		if n == 0 {
			return fmt.Errorf("no host with id %s: %w", uid, store.ErrNotFound)
		}

		if node == nil {
			return nil
		}
		return s.trackEventNode(tx, open, *node, node.UpdatedAt)
	})
}

func (s *SqlStore) ProvisionStatuses(ns *nodeset.NodeSet) (model.ProvisionStatusList, error) {
	var rows []db.ProvisionStatusFetchRow
	if ns == nil {
		all, err := s.q.ProvisionStatusAll(context.Background(), s.ro)
		if err != nil {
			return nil, err
		}
		for _, row := range all {
			rows = append(rows, db.ProvisionStatusFetchRow(row))
		}
	} else {
		some, err := s.q.ProvisionStatusNodeset(context.Background(), s.ro, ns.Iterator().StringSlice())
		if err != nil {
			return nil, err
		}
		for _, row := range some {
			rows = append(rows, db.ProvisionStatusFetchRow(row))
		}
	}

	list := make(model.ProvisionStatusList, 0, len(rows))
	for _, row := range rows {
		status, err := provisionStatusFromRow(row)
		if err != nil {
			return nil, err
		}
		list = append(list, *status)
	}

	return list, nil
}

func provisionStatusFromRow(row db.ProvisionStatusFetchRow) (*model.ProvisionStatus, error) {
	status := &model.ProvisionStatus{
		Node:      row.Name,
		State:     model.ProvisionState(row.State),
		StartedAt: row.StartedAt,
		UpdatedAt: row.UpdatedAt,
	}
	if row.CompletedAt.Valid {
		t := row.CompletedAt.Time
		status.CompletedAt = &t
	}
	if err := json.Unmarshal([]byte(row.Detail), &status.ProvisionStatusDetail); err != nil {
		return nil, fmt.Errorf("bad provision status detail for %s: %w", row.Name, err)
	}

	return status, nil
}
