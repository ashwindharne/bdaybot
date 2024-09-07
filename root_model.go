package main

import (
	"database/sql"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

type RootModel struct {
	model tea.Model
}

func NewRootModel(db *sql.DB) RootModel {
	var rootModel tea.Model
	pnf := EmptyPhoneNumberForm(db)
	rootModel = &pnf
	return RootModel{model: rootModel}
}

func (r RootModel) Init() tea.Cmd {
	return nil
}

func (r RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return r.model.Update(msg)
}

func (r RootModel) View() string {
	return r.model.View()
}

func (r RootModel) Navigate(model tea.Model) (tea.Model, tea.Cmd) {
	r.model = model
	return r.model, r.model.Init()
}
