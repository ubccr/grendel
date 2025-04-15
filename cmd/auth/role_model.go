package auth

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/paginator"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
	p.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "235", Dark: "252"}).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "250", Dark: "238"}).Render("•")

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

		case "left", "h":
			if !m.paginator.OnFirstPage() {
				m.cursor -= m.paginator.PerPage
			}

		case "right", "l":
			if !m.paginator.OnLastPage() {
				m.cursor += m.paginator.PerPage
			}
			if m.cursor > len(m.choices) {
				m.cursor = len(m.choices) - 1
			}

		case "enter", " ":
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

	m.paginator, c = m.paginator.Update(msg)

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

		b.WriteString(fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice))
	}
	b.WriteString("\n" + m.paginator.View() + "\n")
	b.WriteString("k/j: up/down  h/l: left/right \n")
	b.WriteString("q: quit s: save \n")

	return b.String()
}
