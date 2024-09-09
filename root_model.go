package main

import (
	tea "github.com/charmbracelet/bubbletea"
)

type RootModel struct {
	model tea.Model
}

func EmptyRootModel(page tea.Model) RootModel {
	return RootModel{model: page}
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
