package app

import "clashctl-tui/internal/cli"

type App struct {
	cli      cli.Runner
	pages    []Page
	selected int
	width    int
	height   int
	focused  bool
	output   string
	message  string
	running  bool
}

type Page struct {
	Key     string
	Title   string
	Summary string
	Actions []Action
	Preview func(*App) string
}

type Action struct {
	Key   string
	Label string
	Run   func(*App) string
}
