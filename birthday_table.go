package main

import (
	"database/sql"
	"fmt"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	_ "modernc.org/sqlite"
	"strconv"
)

// BIRTHDAY TABLE KEYMAPS
type btKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Create   key.Binding
	Edit     key.Binding
	Settings key.Binding
	Quit     key.Binding
}

func (k btKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Create, k.Edit, k.Settings, k.Quit}
}

func (k btKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Create, k.Edit}, // first column
		{k.Settings, k.Quit},             // second column
	}
}

var btKeys = btKeyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "move down"),
	),
	Create: key.NewBinding(
		key.WithKeys("c"),
		key.WithHelp("c", "create birthday"),
	),
	Edit: key.NewBinding(
		key.WithKeys("e", "enter"),
		key.WithHelp("e", "edit birthday"),
	),
	Settings: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "settings"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c", "q"),
		key.WithHelp("q", "quit"),
	),
}

// BIRTHDAY TABLE MODEL

type BtModel struct {
	phoneNumber string
	table       table.Model
	width       int
	styles      *Styles
	lg          *lipgloss.Renderer
	db          *sql.DB
	help        help.Model
	km          btKeyMap
}

// BIRTHDAY TABLE INITIALIZATION
func EmptyBirthdayTable(phoneNumber string, db *sql.DB) BtModel {
	columns := []table.Column{
		{Title: "ID", Width: 0},
		{Title: "Name", Width: 24},
		{Title: "Birthday", Width: 12},
		{Title: "How Soon?", Width: 16},
	}
	t := table.New(
		table.WithColumns(columns),
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

	// HELP INITIALIZATION
	h := help.New()

	m := BtModel{
		phoneNumber: phoneNumber,
		table:       t,
		db:          db,
		help:        h,
		km:          btKeys,
	}
	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	return m

}

// BIRTHDAY TABLE COMMANDS

type getBirthdaysSuccessMsg struct {
	reminders []birthdayReminder
}

type birthdayReminder struct {
	id    int
	name  string
	month int
	day   int
	year  int
}

func getBirthdays(db *sql.DB, phoneNumber string) tea.Cmd {
	return func() tea.Msg {
		results, err := db.Query(`
	select birthdays.id, name, month, day, year
	from birthdays
	join phone_numbers on phone_numbers.id = birthdays.phone_number_id
	where phone_numbers.phone_number = ?
	ORDER BY
		CASE
			WHEN (month > strftime('%m', 'now') OR (month = strftime('%m', 'now') AND day >= strftime('%d', 'now')))
			THEN date(strftime('%Y', 'now') || '-' || printf('%02d', month) || '-' || printf('%02d', day))
			ELSE date((strftime('%Y', 'now')+1) || '-' || printf('%02d', month) || '-' || printf('%02d', day))
		END; `, phoneNumber)
		if err != nil {
			panic(err)
		}
		defer results.Close()
		reminders := []birthdayReminder{}
		for results.Next() {
			var r = birthdayReminder{}
			err := results.Scan(&r.id, &r.name, &r.month, &r.day, &r.year)
			if err != nil {
				panic(err)
			}
			reminders = append(reminders, r)
		}
		return getBirthdaysSuccessMsg{reminders}
	}
}

// BIRTHDAY TABLE UPDATE-VIEW LOOP

func (m *BtModel) Init() tea.Cmd {
	return getBirthdays(m.db, m.phoneNumber)
}

func (m *BtModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}

func (m *BtModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, 120) - m.styles.Base.GetHorizontalFrameSize()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.km.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.km.Create):
			newForm := EmptyBirthdayForm(m.phoneNumber, m.db)
			return NewRootModel(m.db).Navigate(&newForm)
		case key.Matches(msg, m.km.Edit):
			editingId, err := strconv.Atoi(m.table.SelectedRow()[0])
			if err != nil {
				panic(err)
			}
			editForm := EditBirthdayForm(m.phoneNumber, editingId, m.db)
			return NewRootModel(m.db).Navigate(&editForm)
		}
	case getBirthdaysSuccessMsg:
		var rows []table.Row
		for _, reminder := range msg.reminders {
			rows = append(
				rows,
				[]string{
					strconv.Itoa(reminder.id),
					reminder.name,
					fmt.Sprintf("%d/%d/%d", reminder.month, reminder.day, reminder.year),
					daysTilString(reminder.month, reminder.day),
				},
			)
		}
		m.table.SetRows(rows)
		return m, nil
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *BtModel) View() string {
	header := m.appBoundaryView("Birthday")
	body := baseStyle.Render(m.table.View())
	footer := m.appBoundaryView(m.help.ShortHelpView(m.km.ShortHelp()))
	return header + "\n\n" + body + "\n" + footer
}
