package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"os"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type reminder struct {
	name  string
	month int
	day   int
	year  int
}

type screen int

const (
	PHONE_NUMBER screen = iota
	TABLE
	FORM
)

type model struct {
	currentScreen       screen
	phoneNumber         string
	phoneNumberVerified bool
	phoneNumberForm     *huh.Form
	table               table.Model
	birthdayForm        *huh.Form
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			return m, tea.Quit
		case "enter":
			return m, tea.Batch(
				tea.Printf("Selected %s's birthday!", m.table.SelectedRow()[0]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	switch m.currentScreen {
	case PHONE_NUMBER:
		return m.phoneNumberForm.View()
	case TABLE:
		return baseStyle.Render(m.table.View()) +
			"\n  " + m.table.HelpView() + "\n"
	case FORM:
		return m.birthdayForm.View()
	}
	return "what the fuck how did this happen"
}

func main() {

	phoneNumberForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("phone").Title("Phone Number"),
		),
	)

	columns := []table.Column{
		{Title: "Name", Width: 32},
		{Title: "Birthday", Width: 10},
	}

	rows := []table.Row{
		{"Adrielle Lee", "11/22/1999"},
		{"Ashwin Dharne", "12/26/1998"},
		{"Dan Lam", "06/09/1998"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(16),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(s)

	birthdayForm := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("name").Title("Name"),
			huh.NewSelect[string]().
				Key("month").
				Options(
					huh.NewOptions(
						"January",
						"February",
						"March",
						"April",
						"May",
						"June",
						"July",
						"August",
						"September",
						"October",
						"November",
						"December")...),
		),
	)

	p := tea.NewProgram(model{
		currentScreen:   PHONE_NUMBER,
		phoneNumber:     "",
		phoneNumberForm: phoneNumberForm,
		table:           t,
		birthdayForm:    birthdayForm,
	},
		tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
