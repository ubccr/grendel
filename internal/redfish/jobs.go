// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"strings"

	"github.com/stmcginnis/gofish/schemas"
)

func (s *Service) jobsOne(ctx context.Context, t Target) (*JobList, error) {
	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	js, err := jobService(c)
	if err != nil {
		return nil, err
	}

	jobs, err := js.Jobs()
	if err != nil {
		return nil, err
	}

	return &JobList{Jobs: jobs}, nil
}

func (s *Service) clearJobsOne(ctx context.Context, t Target, req ClearJobsRequest) (*Ack, error) {
	if !req.All && (len(req.IDs) == 0 || slices.Contains(req.IDs, "")) {
		return nil, errors.New("no job named and clear-all not requested")
	}

	c, err := s.connect(ctx, t)
	if err != nil {
		return nil, err
	}
	defer c.Close()

	js, err := jobService(c)
	if err != nil {
		return nil, err
	}

	var ids, uris []string
	if req.All {
		jobs, err := js.Jobs()
		if err != nil {
			return nil, err
		}

		if len(jobs) == 0 {
			return doneWith("no jobs to clear")
		}

		for _, job := range jobs {
			ids = append(ids, job.ID)
		}
	} else {
		ids = req.IDs
	}

	for _, id := range ids {
		uris = append(uris, strings.TrimRight(js.ODataID, "/")+"/Jobs/"+url.PathEscape(id))
	}

	return clearEach("job(s)", ids, func(i int) error {
		_, err := c.client.Delete(uris[i])
		return err
	})
}

// jobService returns a valid redfish JobService handling if the bmc does not implement one
func jobService(c *conn) (*schemas.JobService, error) {
	js, err := c.service.JobService()
	if err != nil {
		return nil, err
	}

	if js == nil {
		return nil, errors.New("bmc does not implement a redfish job service")
	}

	return js, nil
}
