// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package redfish

import (
	"context"
	"errors"
	"net/url"
	"slices"
	"strconv"
	"strings"
	"time"

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

	sortJobsNewestFirst(jobs)

	return &JobList{Jobs: jobs}, nil
}

// sortJobsNewestFirst orders jobs by their latest timestamp, newest first. BMCs list the collection in whatever order they keep it, oldest first on an iDRAC, so the most recent job would otherwise sit at the bottom
func sortJobsNewestFirst(jobs []*schemas.Job) {
	slices.SortStableFunc(jobs, func(a, b *schemas.Job) int {
		if c := jobTime(b).Compare(jobTime(a)); c != 0 {
			return c
		}

		return compareJobIDs(b.ID, a.ID)
	})
}

// jobTime reads a job's timestamp. A job carrying none sorts as the oldest. iDRAC fills StartTime with placeholders like TIME_NOW, which fail to parse and fall through to the next field
func jobTime(j *schemas.Job) time.Time {
	for _, raw := range []string{j.CreationTime, j.StartTime, j.EndTime} {
		if ts, err := time.Parse(time.RFC3339, raw); err == nil {
			return ts
		}
	}

	return time.Time{}
}

// compareJobIDs compares IDs by their trailing number when both have one, so JID_10 comes after JID_9. iDRAC job IDs encode when the job was created, which keeps the order right when the timestamps are missing
func compareJobIDs(a, b string) int {
	na, oka := trailingNumber(a)
	nb, okb := trailingNumber(b)
	if oka && okb && na != nb {
		if na < nb {
			return -1
		}
		return 1
	}

	return strings.Compare(a, b)
}

func trailingNumber(id string) (uint64, bool) {
	i := len(id)
	for i > 0 && id[i-1] >= '0' && id[i-1] <= '9' {
		i--
	}

	n, err := strconv.ParseUint(id[i:], 10, 64)
	return n, err == nil
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
