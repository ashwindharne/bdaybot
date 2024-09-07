package main

import (
	"database/sql"
	"fmt"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"slices"
	"strconv"
	"time"
)

type bfKeyMap struct {
	Back key.Binding
	Quit key.Binding
}

func (k bfKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Back, k.Quit}
}
func (k bfKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{
		k.Back,
	}}
}

var bfKeys = bfKeyMap{
	Back: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
}

type bfState struct {
	phoneNumber string
	name        string
	month       int
	day         string
	year        string
	editingId   int
}

type BfModel struct {
	state  bfState
	form   *huh.Form
	width  int
	styles *Styles
	lg     *lipgloss.Renderer
	db     *sql.DB
	km     bfKeyMap
}

func validateDay(day string) error {
	if day == "" {
		return fmt.Errorf("day must be number between 1 and 31")
	}
	dayInt, err := strconv.Atoi(day)
	if err != nil {
		return fmt.Errorf("day must be number between 1 and 31")
	}
	if dayInt < 1 || dayInt > 31 {
		return fmt.Errorf("day must be number between 1 and 31")
	}
	return nil
}

func validateYear(year string) error {
	thisYear, _, _ := time.Now().Date()
	if year == "" {
		return fmt.Errorf("year must be number between 1 and %d", thisYear)
	}
	yearInt, err := strconv.Atoi(year)
	if err != nil {
		return fmt.Errorf("year must be number between 1 and %d", thisYear)
	}
	if yearInt < 1 || yearInt > thisYear {
		return fmt.Errorf("year must be number between 1 and %d", thisYear)
	}
	return nil
}

func PopulatedBF(initState bfState, db *sql.DB) BfModel {
	bf := BfModel{
		state: initState,
		db:    db,
		lg:    lipgloss.DefaultRenderer(),
		km:    bfKeys,
	}
	bf.styles = NewStyles(bf.lg)
	bf.form = huh.NewForm(
		huh.NewGroup(
			huh.NewInput().
				Key("name").
				Title("Name").
				Description("Enter the name of the person whose birthday you'd like to be reminded of").
				Value(&bf.state.name),
			huh.NewSelect[int]().
				Key("month").
				Title("Month").
				Options(
					huh.NewOption("January", 1),
					huh.NewOption("February", 2),
					huh.NewOption("March", 3),
					huh.NewOption("April", 4),
					huh.NewOption("May", 5),
					huh.NewOption("June", 6),
					huh.NewOption("July", 7),
					huh.NewOption("August", 8),
					huh.NewOption("September", 9),
					huh.NewOption("October", 10),
					huh.NewOption("November", 11),
					huh.NewOption("December", 12),
				).
				Value(&bf.state.month).
				Description("Enter the month of their birthday."),
			huh.NewInput().
				Title("Day").
				Description("Enter the day of their birthday.").
				Key("day").
				Value(&bf.state.day).
				CharLimit(2).
				Validate(validateDay),
			huh.NewInput().
				Title("Year").
				Key("year").
				Description("Enter the year of their birthday.").
				Value(&bf.state.year).
				CharLimit(4).
				Validate(validateYear),
			huh.NewConfirm().
				Key("confirm").
				Title("Save Changes?").
				Affirmative("Yep").
				Negative("Nope"),
		),
	)
	return bf
}

func InitBF(phoneNumber string, db *sql.DB) BfModel {
	initState := bfState{
		phoneNumber: phoneNumber,
		day:         "1",
		month:       1,
		year:        "1",
	}
	return PopulatedBF(initState, db)
}

func EditBF(phoneNumber string, editingId int, db *sql.DB) BfModel {
	initState := bfState{
		phoneNumber: phoneNumber,
	}
	day, year := 0, 0
	err := db.QueryRow(`
SELECT id, name, month, day, year
FROM birthdays
WHERE id = ?
`, editingId).Scan(&initState.editingId, &initState.name, &initState.month, &day, &year)
	if err != nil {
		panic(err)
	}
	initState.day = strconv.Itoa(day)
	initState.year = strconv.Itoa(year)
	return PopulatedBF(initState, db)
}

func (m *BfModel) Init() tea.Cmd {
	return m.form.PrevField()
}

func submitChanges(m *BfModel) {
	if m.form.GetBool("confirm") {
		dayStr, yearStr := m.form.GetString("day"), m.form.GetString("year")
		day, err1 := strconv.Atoi(dayStr)
		year, err2 := strconv.Atoi(yearStr)
		if err1 != nil {
			panic(err1)
		}
		if err2 != nil {
			panic(err2)
		}
		if m.state.editingId == 0 {
			_, err := m.db.Exec(`
INSERT INTO BIRTHDAYS 
(phone_number_id, name, month, day, year)
values (
(select id from phone_numbers where phone_number = ?),
?, ?, ?, ?
)
`, m.state.phoneNumber, m.form.Get("name"), m.form.GetInt("month"), day, year)
			if err != nil {
				panic(err)
			}
		} else {
			_, err := m.db.Exec(`
UPDATE birthdays
SET name = ?, month = ?, day = ?, year = ?
WHERE id = ?
`, m.form.Get("name"), m.form.GetInt("month"), day, year, m.state.editingId)
			if err != nil {
				panic(err)
			}
		}
	}
}

func (m *BfModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, 120) - m.styles.Base.GetHorizontalFrameSize()
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.km.Quit):
			return m, tea.Quit
		case key.Matches(msg, m.km.Back):
			bt := InitBT(m.state.phoneNumber, m.db)
			return NewRootModel(m.db).Navigate(&bt)
		}
	}
	f, cmd := m.form.Update(msg)
	m.form = f.(*huh.Form)

	if m.form.State == huh.StateCompleted {
		submitChanges(m)
		bt := InitBT(m.state.phoneNumber, m.db)
		return NewRootModel(m.db).Navigate(&bt)
	}
	return m, cmd
}

func (m *BfModel) View() string {
	header := m.appBoundaryView("Birthday Bot")
	body := baseStyle.Render(m.form.WithShowHelp(false).View())
	footer := m.appBoundaryView(m.form.Help().ShortHelpView(slices.Concat(m.km.ShortHelp(), m.form.KeyBinds())))
	return header + "\n" + body + "\n" + footer
}

func (m *BfModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(indigo),
	)
}
