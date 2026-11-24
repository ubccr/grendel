// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"net/http"
	"slices"
	"strings"

	"github.com/go-fuego/fuego"
	"github.com/ubccr/grendel/internal/settings"
	"github.com/ubccr/grendel/pkg/model"
)

type PatchSettingsRequest struct {
	// Set maps a setting's key to its new value: either of the setting's own JSON type or as text, the way it would be typed on the command line
	Set map[string]any `json:"set"`
	// Unset lists keys to remove from the database, which returns them to their default
	Unset []string `json:"unset"`
}

func (h *Handler) GetSettings(ctx fuego.ContextNoBody) (*settings.Listing, error) {
	list := settings.List()
	return &list, nil
}

func (h *Handler) PatchSettings(ctx fuego.ContextWithBody[PatchSettingsRequest]) (*settings.Listing, error) {
	body, err := ctx.Body()
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to parse body"}
	}

	refused, err := settings.Update(h.DB, body.Set, body.Unset, actor(ctx.Context()))
	if err != nil {
		return nil, fuego.HTTPError{Err: err, Title: "Error", Detail: "failed to save settings"}
	}
	if len(refused) > 0 {
		return nil, refusedError(refused)
	}

	keys := slices.Sorted(slices.Values(append(slices.Collect(maps.Keys(body.Set)), body.Unset...)))
	if len(keys) > 0 {
		h.audit(ctx.Context(), model.SeverityInfo, "settings.update", fmt.Sprintf("Changed %d settings", len(keys)), strings.Join(keys, ", "))
	}

	list := settings.List()
	return &list, nil
}

func refusedError(refused map[string]string) error {
	items := make([]fuego.ErrorItem, 0, len(refused))
	for key, reason := range refused {
		items = append(items, fuego.ErrorItem{Name: key, Reason: reason})
	}
	slices.SortFunc(items, func(a, b fuego.ErrorItem) int { return strings.Compare(a.Name, b.Name) })
	return fuego.HTTPError{
		Err:    errors.New("settings refused"),
		Status: http.StatusUnprocessableEntity,
		Title:  "Invalid settings",
		Detail: "nothing was saved",
		Errors: items,
	}
}

// restoreSettings stores the settings of a dump on top of the ones already there
func (h *Handler) restoreSettings(stored []model.StoredSetting, who string) error {
	rows := make([]settings.Row, len(stored))
	for i, s := range stored {
		rows[i] = settings.Row{Key: s.Key, Value: json.RawMessage(s.Value)}
	}
	skipped, err := settings.Restore(h.DB, rows, who)
	for _, key := range slices.Sorted(maps.Keys(skipped)) {
		log.Warnf("not restoring %s: %s", key, skipped[key])
	}
	return err
}
