package main

import (
	"database/sql"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"strconv"
)

type pnfState struct {
	phoneNumber string
}

type PnfModel struct {
	state  pnfState
	form   *huh.Form
	width  int
	styles *Styles
	lg     *lipgloss.Renderer
	db     *sql.DB
}

func validateUSPhoneNumber(phoneNumber string) error {
	if len(phoneNumber) < 12 || len(phoneNumber) > 12 {
		return fmt.Errorf("must be exactly 12 digits")
	}

	if phoneNumber[0:2] != "+1" {
		return fmt.Errorf("not a valid US phone number")
	}
	if _, err := strconv.ParseInt(phoneNumber[2:], 10, 64); err != nil {
		return fmt.Errorf("numbers only")
	}
	return nil
}

func InitPNF(db *sql.DB) PnfModel {
	m := PnfModel{
		state: pnfState{
			phoneNumber: "+1",
		},
		db: db,
		lg: lipgloss.DefaultRenderer(),
	}
	m.styles = NewStyles(m.lg)
	f := huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("phone").
				Title("Enter your phone number.").
				Description("Please enter the phone number that you would like alerts to be sent to.").
				Validate(validateUSPhoneNumber).
				Value(&m.state.phoneNumber),
		),
	).WithShowHelp(false)
	f.PrevGroup()
	m.form = f
	return m
}

func (m *PnfModel) Init() tea.Cmd {
	return nil
}

func (m *PnfModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, 120) - m.styles.Base.GetHorizontalFrameSize()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}
	}
	f, cmd := m.form.Update(msg)
	m.form = f.(*huh.Form)
	phoneNumber := m.form.GetString("phone")
	if m.form.State == huh.StateCompleted {
		m.db.Exec(`
INSERT OR IGNORE INTO phone_numbers (phone_number, verified)
values (?, TRUE);
`, phoneNumber)
		bt := InitBT(phoneNumber, m.db)
		return NewRootModel(m.db).Navigate(&bt)
	}
	return m, cmd
}

func (m *PnfModel) View() string {
	header := m.appBoundaryView("Birthday Bot")
	body := m.form.View()
	footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
	return header + "\n" + body + "\n" + footer
}

func (m *PnfModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}
