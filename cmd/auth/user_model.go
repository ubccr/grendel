package auth

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type userAddModel struct {
	focusIndex int
	inputs     []textinput.Model
	save       bool
}

func initialUserAddModel() userAddModel {
	m := userAddModel{
		inputs: make([]textinput.Model, 2),
		save:   false,
	}
	var t textinput.Model

	for i := range m.inputs {
		t = textinput.New()
		t.CharLimit = 256
		t.Width = 40

		switch i {
		case 0:
			t.Placeholder = "Password"
			t.Focus()
			t.EchoMode = textinput.EchoPassword
		case 1:
			t.Placeholder = "Confirm Password"
			t.EchoMode = textinput.EchoPassword
		}

		m.inputs[i] = t
	}

	return m
}

func (m userAddModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m userAddModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "up", "down", "enter":
			s := msg.String()

			if s == "up" {
				m.focusIndex--
			}

			if m.focusIndex == len(m.inputs)-1 && s == "enter" {
				m.save = true
				return m, tea.Quit
			}

			if s == "down" || s == "enter" {
				m.focusIndex++
			}

			if m.focusIndex > len(m.inputs)-1 {
				m.focusIndex = 0
			} else if m.focusIndex < 0 {
				m.focusIndex = len(m.inputs)
			}

			cmds := make([]tea.Cmd, len(m.inputs))
			for i := 0; i <= len(m.inputs)-1; i++ {
				if i == m.focusIndex {
					cmds[i] = m.inputs[i].Focus()
					continue
				}
				m.inputs[i].Blur()
			}

			return m, tea.Batch(cmds...)
		}
	}
	cmds := make([]tea.Cmd, len(m.inputs))
	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}

	return m, tea.Batch(cmds...)
}

func (m userAddModel) View() string {
	var b strings.Builder

	b.WriteString("Enter users password:\n")
	for i := range m.inputs {
		b.WriteString(m.inputs[i].View())
		if i < len(m.inputs)-1 {
			b.WriteRune('\n')
		}
	}
	b.WriteString("\n\nq/ctrl+c: quit  enter: submit\n")

	return b.String()
}
