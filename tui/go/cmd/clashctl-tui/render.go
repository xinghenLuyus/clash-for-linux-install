package main

import (
	"fmt"
	"strings"

	"clashctl/internal/ui"
)

func (a *app) render() {
	var b strings.Builder
	b.Grow(a.width * a.height)
	if a.width != a.lastWidth || a.height != a.lastHeight {
		b.WriteString("\033[2J")
		a.lastWidth = a.width
		a.lastHeight = a.height
	}
	b.WriteString("\033[H")
	top := a.statusCache
	if top == "" {
		top = " clashctl TUI | loading "
	}
	b.WriteString(ui.Invert(ui.Pad(top, a.width)))
	b.WriteString("\033[K\r\n")
	leftW := 24
	rightW := a.width - leftW - 3
	bodyH := a.height - 3
	content := ui.WrapLines(a.contentCache, rightW)
	for row := 0; row < bodyH; row++ {
		left := ""
		if row < len(a.pages) {
			p := a.pages[row]
			prefix := "  "
			if row == a.selected {
				prefix = "> "
			}
			left = prefix + p.title
			if a.focused && row == a.selected {
				left += " >"
			}
		}
		right := ""
		if row < len(content) {
			right = content[row]
		}
		b.WriteString(ui.Pad(left, leftW))
		b.WriteString(" │ ")
		b.WriteString(ui.Pad(right, rightW))
		b.WriteString("\033[K\r\n")
	}
	footer := a.footer()
	if a.message != "" {
		footer += " | " + a.message
	}
	b.WriteString(ui.Invert(ui.Pad(footer, a.width)))
	b.WriteString("\033[K")
	fmt.Print(b.String())
}
