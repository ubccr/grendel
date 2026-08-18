// SPDX-FileCopyrightText: (C) 2019 Grendel Authors
//
// SPDX-License-Identifier: GPL-3.0-or-later

package auth

import (
	"fmt"
	"os"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	role      string
	choices   []string
	cursor    int
	selected  map[int]struct{}
	paginator paginator.Model
	save      bool
}

func InitialModel(role string) model {
	choices := []string{}
	selected := make(map[int]struct{})

	p := paginator.New()
	p.Type = paginator.Dots
	p.PerPage = 10
	// lipgloss v2 dropped AdaptiveColor, background detection is now explicit
	lightDark := lipgloss.LightDark(lipgloss.HasDarkBackground(os.Stdin, os.Stdout))
	p.ActiveDot = lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("235"), lipgloss.Color("252"))).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lightDark(lipgloss.Color("250"), lipgloss.Color("238"))).Render("•")

	return model{
		role:      role,
		choices:   choices,
		selected:  selected,
		paginator: p,
		save:      false,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var c tea.Cmd
	switch msg := msg.(type) {

	case tea.KeyMsg:
		start, end := m.paginator.GetSliceBounds(len(m.choices))

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > start {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < end-1 {
				m.cursor++
			}

		case "enter", " ":
			if m.cursor < 0 || m.cursor >= len(m.choices) {
				break
			}
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
			}

		case "s":
			m.save = true
			return m, tea.Quit
		}

	}

	// the paginator turns the page on its own keys, left/h/pgup and right/l/pgdown, so the cursor is held on the same row of whichever page it lands on rather than being left behind on the old one
	start, _ := m.paginator.GetSliceBounds(len(m.choices))
	offset := m.cursor - start

	m.paginator, c = m.paginator.Update(msg)

	start, end := m.paginator.GetSliceBounds(len(m.choices))
	m.cursor = min(max(start+offset, start), end-1)

	return m, c
}

func (m model) View() string {
	var b strings.Builder
	b.WriteString("Add or remove permissions from the role:\n\n")

	start, end := m.paginator.GetSliceBounds(len(m.choices))
	for i, choice := range m.choices[start:end] {
		cursor := " "
		if m.cursor == start+i {
			cursor = ">"
		}

		checked := " "
		if _, ok := m.selected[start+i]; ok {
			checked = "x"
		}

		fmt.Fprintf(&b, "%s [%s] %s\n", cursor, checked, choice)
	}
	fmt.Fprintf(&b, "\n%s\n", m.paginator.View())
	b.WriteString("k/j: up/down  h/l: left/right \n")
	b.WriteString("q: quit s: save \n")

	return b.String()
}
