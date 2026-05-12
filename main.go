package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
)

const (
	host = "0.0.0.0"
	port = "2222"
)

// Styles
var (
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.AdaptiveColor{Light: "#04B575", Dark: "#04B575"}).
				Render
)

// Item represents a menu item
type item struct {
	title       string
	description string
}

func (i item) Title() string       { return i.title }
func (i item) Description() string { return i.description }
func (i item) FilterValue() string { return i.title }

// Model is the main application model
type model struct {
	list     list.Model
	choice   string
	quitting bool
	width    int
	height   int
}

func initialModel() model {
	items := []list.Item{
		item{title: "Dashboard", description: "View system dashboard"},
		item{title: "Projects", description: "Browse your projects"},
		item{title: "Settings", description: "Configure your preferences"},
		item{title: "About", description: "Learn more about this app"},
	}

	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#25A065")).
		BorderLeftForeground(lipgloss.Color("#25A065"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#25A065")).
		BorderLeftForeground(lipgloss.Color("#25A065"))

	l := list.New(items, delegate, 0, 0)
	l.Title = "SSH TUI App"
	l.Styles.Title = titleStyle

	return model{list: l}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetSize(msg.Width-4, msg.Height-4)
		return m, nil

	case tea.KeyMsg:
		if m.choice != "" {
			if msg.String() == "esc" || msg.String() == "backspace" {
				m.choice = ""
				return m, nil
			}
		}

		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if i, ok := m.list.SelectedItem().(item); ok {
				m.choice = i.title
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Thanks for visiting! Goodbye.\n"
	}

	if m.choice != "" {
		return appStyle.Render(m.detailView())
	}

	return appStyle.Render(m.list.View())
}

func (m model) detailView() string {
	titleBar := titleStyle.Render(m.choice)

	var content string
	switch m.choice {
	case "Dashboard":
		content = m.dashboardView()
	case "Projects":
		content = m.projectsView()
	case "Settings":
		content = m.settingsView()
	case "About":
		content = m.aboutView()
	default:
		content = "Unknown section"
	}

	help := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Render("\n  esc/backspace: back  q: quit")

	return fmt.Sprintf("%s\n\n%s\n%s", titleBar, content, help)
}

func (m model) dashboardView() string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#25A065")).
		Padding(1, 3).
		Width(40)

	stats := fmt.Sprintf(
		"  Active Users:  %s\n  Uptime:        %s\n  Memory:        %s\n  CPU:           %s",
		lipgloss.NewStyle().Bold(true).Render("42"),
		lipgloss.NewStyle().Bold(true).Render("99.9%"),
		lipgloss.NewStyle().Bold(true).Render("2.1 GB"),
		lipgloss.NewStyle().Bold(true).Render("23%"),
	)

	return box.Render(stats)
}

func (m model) projectsView() string {
	projects := []struct{ name, status string }{
		{"api-server", "running"},
		{"web-frontend", "running"},
		{"worker-queue", "stopped"},
		{"ml-pipeline", "deploying"},
	}

	style := lipgloss.NewStyle().PaddingLeft(2)
	var s string
	for _, p := range projects {
		statusColor := "#FF0000"
		if p.status == "running" {
			statusColor = "#04B575"
		} else if p.status == "deploying" {
			statusColor = "#FFD700"
		}
		dot := lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Render("●")
		s += fmt.Sprintf("  %s  %-20s %s\n", dot, p.name, p.status)
	}

	return style.Render(s)
}

func (m model) settingsView() string {
	return lipgloss.NewStyle().PaddingLeft(2).Render(
		"  Theme:         dark\n  Notifications: enabled\n  Language:      en-US\n  Timezone:      UTC",
	)
}

func (m model) aboutView() string {
	banner := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#25A065")).
		Bold(true).
		Render("SSH TUI App v0.1.0")

	desc := "A terminal application accessible via SSH.\nBuilt with Wish + Bubble Tea + Lip Gloss."

	return fmt.Sprintf("  %s\n\n  %s", banner, desc)
}

func main() {
	s, err := wish.NewServer(
		wish.WithAddress(net.JoinHostPort(host, port)),
		wish.WithHostKeyPath(".ssh/id_ed25519"),
		wish.WithMiddleware(
			bubbletea.Middleware(teaHandler),
			activeterm.Middleware(),
		),
	)
	if err != nil {
		log.Fatalln("Could not create server:", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("Starting SSH server on %s:%s", host, port)
	log.Printf("Connect with: ssh localhost -p %s", port)

	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Fatalln("Server error:", err)
		}
	}()

	<-done
	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := s.Shutdown(ctx); err != nil {
		log.Fatalln("Shutdown error:", err)
	}
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, _ := s.Pty()
	m := initialModel()
	m.width = pty.Window.Width
	m.height = pty.Window.Height
	m.list.SetSize(pty.Window.Width-4, pty.Window.Height-4)
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}
