package main

import (
	"context"
	"database/sql"
	"errors"
	"flag"
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	dbPath = "db.sqlite"
)

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	// This should never fail, as we are using the activeterm middleware.
	//pty, _, _ := s.Pty()

	// When running a Bubble Tea app over SSH, you shouldn't use the default
	// lipgloss.NewStyle function.
	// That function will use the color profile from the os.Stdin, which is the
	// server, not the client.
	// We provide a MakeRenderer function in the bubbletea middleware package,
	// so you can easily get the correct renderer for the current session, and
	// use it to create the styles.
	// The recommended way to use these styles is to then pass them down to
	// your Bubble Tea model.
	renderer := bubbletea.MakeRenderer(s)

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	pnf := EmptyPhoneNumberForm(db, renderer, NewStyles(renderer))
	return EmptyRootModel(&pnf), []tea.ProgramOption{tea.WithAltScreen()}
}

const (
	host = "0.0.0.0"
	port = "23234"
)

func runWishServer(dbPath string) {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(), // Bubble Tea apps usually require a PTY.
			logging.Middleware(),
		),
	)
	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}

func runApp(dbPath string) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		panic(err)
	}
	renderer := lipgloss.DefaultRenderer()
	styles := NewStyles(renderer)
	pnf := EmptyPhoneNumberForm(db, renderer, styles)
	p := tea.NewProgram(EmptyRootModel(&pnf), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

func main() {
	dbPathPtr := flag.String("db", "db.sqlite", "path to sqlite database")
	serverPtr := flag.Bool("server", false, "run as SSH server")
	flag.Parse()
	if *serverPtr {
		runWishServer(*dbPathPtr)
	} else {
		runApp(*dbPathPtr)
	}
}
