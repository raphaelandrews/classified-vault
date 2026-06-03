package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"classified-vault/tui"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080", "Backend server URL")
	flag.Parse()

	if envURL := os.Getenv("SERVER_URL"); envURL != "" && *serverURL == "http://localhost:8080" {
		*serverURL = envURL
	}

	p := tea.NewProgram(tui.New(*serverURL), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
