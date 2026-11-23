// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package bmc

import "testing"

func TestEscapeFilter(t *testing.T) {
	for _, tc := range []struct {
		name   string
		reads  []string
		want   string
		action escapeAction
	}{
		{"plain input", []string{"ls -l\r"}, "ls -l\r", escapeNone},
		{"disconnect at the start", []string{"~."}, "", escapeQuit},
		{"disconnect after a line", []string{"ls\r~."}, "ls\r", escapeQuit},
		{"a ~ mid line is sent", []string{"cd ~.ssh"}, "cd ~.ssh", escapeNone},
		{"~~ sends one ~", []string{"~~."}, "~.", escapeNone},
		{"an unknown escape passes through", []string{"~x"}, "~x", escapeNone},
		{"help", []string{"~?"}, "", escapeHelp},
		{"split across reads", []string{"\r", "~", "."}, "\r", escapeQuit},
		{"input pasted after help is still sent", []string{"~?uptime\r"}, "uptime\r", escapeHelp},
		{"nothing is sent after a disconnect", []string{"~.uptime\r"}, "", escapeQuit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := &escapeFilter{lineStart: true}
			var got []byte
			action := escapeNone
		reads:
			for _, r := range tc.reads {
				for in := []byte(r); len(in) > 0; {
					out, a, rest := f.filter(in)
					in = rest
					got = append(got, out...)
					if a != escapeNone {
						action = a
					}
					if a == escapeQuit {
						break reads
					}
				}
			}
			if string(got) != tc.want || action != tc.action {
				t.Errorf("got %q action %d, want %q action %d", got, action, tc.want, tc.action)
			}
		})
	}
}
