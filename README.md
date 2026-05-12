# SSH TUI App

A terminal user interface application accessible via SSH, built with [Wish](https://github.com/charmbracelet/wish), [Bubble Tea](https://github.com/charmbracelet/bubbletea), and [Lip Gloss](https://github.com/charmbracelet/lipgloss).

## Prerequisites

- [Go](https://go.dev/) 1.25.9 or later

## Running the server

```bash
go run main.go
```

The SSH server will start on `0.0.0.0:2222`.

## Connecting to the application

```bash
ssh localhost -p 2222
```

## Navigation

- **Arrow keys** — navigate the menu
- **Enter** — select an item
- **Esc / Backspace** — go back
- **q / Ctrl+C** — quit

## Menu options

- **Dashboard** — View system stats (users, uptime, memory, CPU)
- **Projects** — Browse projects and their status
- **Settings** — View configuration preferences
- **About** — App version and info
