package main

import (
	"database/sql"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"os"
)

func main() {
	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "db.sqlite"
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	p := tea.NewProgram(NewRootModel(db),
		tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
