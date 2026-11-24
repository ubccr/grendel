// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package render

import (
	"bytes"
	"context"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/coreos/butane/config"
	"github.com/coreos/butane/config/common"
	"github.com/fsnotify/fsnotify"
	"github.com/ubccr/grendel/internal/settings"
)

//go:embed templates/ipxe.tmpl
var ipxeTmpl string

//go:embed templates/kickstart.tmpl
var kickstartTmpl string

//go:embed templates/user-data.tmpl
var userDataTmpl string

//go:embed templates/meta-data.tmpl
var metaDataTmpl string

//go:embed templates/butane.tmpl
var butaneTmpl string

// RenderString executes text as a template with the provision template functions, returning text unchanged if it contains no actions
func RenderString(name, text string, data map[string]any) (string, error) {
	if !strings.Contains(text, "{{") {
		return text, nil
	}

	tmpl, err := template.New(name).Funcs(funcMap).Option("missingkey=error").Parse(text)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Set is the built-in templates plus those in provision.templates_dir
type Set struct {
	templates atomic.Pointer[template.Template]
	// quiet leaves what was found out of the log
	quiet bool
}

func NewSet() (*Set, error) {
	s := &Set{}
	return s, s.reload()
}

// ReadSet reads the templates as NewSet does without logging what it found, for a caller that reads them on every request
func ReadSet() (*Set, error) {
	s := &Set{quiet: true}
	return s, s.reload()
}

func buildTemplates(quiet bool) (*template.Template, error) {
	tmpl, err := template.New("ipxe.tmpl").Funcs(funcMap).Parse(ipxeTmpl)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.New("kickstart.tmpl").Funcs(funcMap).Parse(kickstartTmpl)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.New("user-data.tmpl").Funcs(funcMap).Parse(userDataTmpl)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.New("meta-data.tmpl").Funcs(funcMap).Parse(metaDataTmpl)
	if err != nil {
		return nil, err
	}

	tmpl, err = tmpl.New("butane.tmpl").Funcs(funcMap).Parse(butaneTmpl)
	if err != nil {
		return nil, err
	}

	if templates := settings.ProvisionTemplatesDir.Get(); templates != "" {
		dir := filepath.Join(templates, "*.tmpl")
		matches, err := filepath.Glob(dir)
		if err != nil {
			return nil, err
		}

		if len(matches) == 0 {
			if !quiet {
				log.WithField("dir", dir).Warn("no additional templates found in provision.templates_dir")
			}
		} else {
			tmpl, err = tmpl.Funcs(funcMap).ParseGlob(dir)
			if err != nil {
				return nil, err
			}
		}
	}

	if !quiet {
		log.Debug(tmpl.DefinedTemplates())
	}
	return tmpl, nil
}

func (s *Set) reload() error {
	tmpl, err := buildTemplates(s.quiet)
	if err != nil {
		return err
	}

	s.templates.Store(tmpl)
	return nil
}

func (s *Set) Watch(ctx context.Context) error {
	if settings.ProvisionTemplatesDir.Get() == "" {
		return nil
	}

	w, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	err = w.Add(settings.ProvisionTemplatesDir.Get())
	if err != nil {
		return fmt.Errorf("failed to monitor template directory: %w", err)
	}

	go func() {
		defer w.Close()

		var debounce *time.Timer
		var debounceChan <-chan time.Time

		for {
			select {
			case <-ctx.Done():
				return

			case event, ok := <-w.Events:
				if !ok {
					return
				}

				if filepath.Ext(event.Name) != ".tmpl" {
					continue
				}
				if !event.Has(fsnotify.Write | fsnotify.Create | fsnotify.Remove | fsnotify.Rename) {
					continue
				}

				if debounce == nil {
					debounce = time.NewTimer(300 * time.Millisecond)
					debounceChan = debounce.C
					continue
				}

				debounce.Reset(300 * time.Millisecond)

			case <-debounceChan:
				err := s.reload()
				if err != nil {
					log.WithError(err).Error("template hot reload failed, serving previous templates")
					continue
				}

				log.Info("Successfully hot reloaded templates")

			case err, ok := <-w.Errors:
				if !ok {
					return
				}
				log.WithError(err).Warn("failed to watch template directory")
			}
		}
	}()

	return nil
}

func (s *Set) Execute(w io.Writer, name string, data any) error {
	tmpl := s.templates.Load()
	if tmpl == nil {
		return errors.New("failed to load templates")
	}
	return tmpl.ExecuteTemplate(w, name, data)
}

func (s *Set) Has(name string) bool {
	tmpl := s.templates.Load()
	return tmpl != nil && tmpl.Lookup(name) != nil
}

// ErrButane marks a rendered template that Butane would not translate to an Ignition config
var ErrButane = errors.New("failed to translate butane config")

// Ignition renders a Butane template and translates the result to an Ignition config
func (s *Set) Ignition(name string, data any) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := s.Execute(buf, name, data); err != nil {
		return nil, err
	}

	out, report, err := config.TranslateBytes(buf.Bytes(), common.TranslateBytesOptions{Pretty: false})
	details := strings.TrimSpace(report.String())
	if err != nil {
		if details != "" {
			err = fmt.Errorf("%w: %s", err, details)
		}
		return nil, fmt.Errorf("%w: template %s: %w", ErrButane, name, err)
	}

	// Strict like butane --strict, as Butane only warns about unknown keys and the node would boot without that config
	if len(report.Entries) > 0 {
		return nil, fmt.Errorf("%w: template %s: butane warnings: %s", ErrButane, name, details)
	}

	return out, nil
}
