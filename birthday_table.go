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
	"time"
)

// BIRTHDAY TABLE MODEL
type btKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	Create   key.Binding
	Edit     key.Binding
	Settings key.Binding
	Quit     key.Binding
}

// ShortHelp returns keybindings to be shown in the mini help view. It's part
// of the key.Map interface.
func (k btKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Create, k.Edit, k.Settings, k.Quit}
}

// FullHelp returns keybindings for the expanded help view. It's part of the
// key.Map interface.
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

func daysToNextBirthday(bYear int, bMonth int, bDay int) int {
	now := time.Now()
	nYear, _, _ := now.Date()
	birthdayThisYear := time.Date(nYear, time.Month(bMonth), bDay, 0, 0, 0, 0, time.UTC)
	birthdayNextYear := time.Date(nYear+1, time.Month(bMonth), bDay, 0, 0, 0, 0, time.UTC)
	if birthdayThisYear.After(now) {
		return int(birthdayThisYear.Sub(now).Hours() / 24)
	} else {
		return int(birthdayNextYear.Sub(now).Hours() / 24)
	}
}

func daysTilString(bYear int, bMonth int, bDay int) string {
	days := daysToNextBirthday(bYear, bMonth, bDay)
	if days == 0 {
		return "It's today!"
	} else if days == 1 {
		return "It's tomorrow!"
	} else {
		return fmt.Sprintf("%d days", days)
	}
}

func InitBT(phoneNumber string, db *sql.DB) BtModel {
	columns := []table.Column{
		{Title: "ID", Width: 0},
		{Title: "Name", Width: 24},
		{Title: "Birthday", Width: 12},
		{Title: "How Soon?", Width: 16},
	}

	getBirthdaysQuery := `
	select birthdays.id, name, month, day, year
	from birthdays
	join phone_numbers on phone_numbers.id = birthdays.phone_number_id
	where phone_numbers.phone_number = ?
	ORDER BY 
		CASE 
			WHEN (month > strftime('%m', 'now') OR (month = strftime('%m', 'now') AND day >= strftime('%d', 'now')))
			THEN date(strftime('%Y', 'now') || '-' || printf('%02d', month) || '-' || printf('%02d', day))
			ELSE date((strftime('%Y', 'now')+1) || '-' || printf('%02d', month) || '-' || printf('%02d', day))
		END; `
	results, err := db.Query(getBirthdaysQuery, phoneNumber)
	if err != nil {
		panic(err)
	}
	defer results.Close()

	var rows []table.Row
	for results.Next() {
		var id int
		var name string
		var month, day, year int
		err := results.Scan(&id, &name, &month, &day, &year)
		if err != nil {
			panic(err)
		}
		rows = append(rows, []string{strconv.Itoa(id), name, fmt.Sprintf("%d/%d/%d", month, day, year), daysTilString(year, month, day)})
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

func (m *BtModel) Init() tea.Cmd {
	return nil
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
			newForm := InitBF(m.phoneNumber, m.db)
			return NewRootModel(m.db).Navigate(&newForm)
		case key.Matches(msg, m.km.Edit):
			editingId, err := strconv.Atoi(m.table.SelectedRow()[0])
			if err != nil {
				panic(err)
			}
			editForm := EditBF(m.phoneNumber, editingId, m.db)
			return NewRootModel(m.db).Navigate(&editForm)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m *BtModel) View() string {
	header := m.appBoundaryView("Birthday Bot")
	body := baseStyle.Render(m.table.View())
	footer := m.appBoundaryView(m.help.ShortHelpView(m.km.ShortHelp()))
	return header + "\n" + body + "\n" + footer
}
