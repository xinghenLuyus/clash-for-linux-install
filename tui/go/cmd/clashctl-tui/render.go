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
	if a.modal != nil {
		a.renderModal(&b)
	}
	fmt.Print(b.String())
}

func (a *app) renderModal(b *strings.Builder) {
	w := a.width * 3 / 5
	if w < 52 {
		w = 52
	}
	if w > a.width-8 {
		w = a.width - 8
	}
	h := 9
	x := (a.width - w) / 2
	y := (a.height - h) / 2
	if x < 1 {
		x = 1
	}
	if y < 2 {
		y = 2
	}
	line := func(row int, text string) {
		b.WriteString(fmt.Sprintf("\033[%d;%dH", y+row, x+1))
		b.WriteString(ui.Invert(ui.Pad(text, w)))
	}
	line(0, "┌"+strings.Repeat("─", w-2)+"┐")
	line(1, "│ "+ui.Pad(a.modal.Title, w-4)+" │")
	line(2, "├"+strings.Repeat("─", w-2)+"┤")
	line(3, "│ "+ui.Pad(a.modal.Label, w-4)+" │")
	line(4, "│ "+ui.Pad(a.modal.Value, w-4)+" │")
	errText := a.modal.Error
	if errText == "" {
		errText = "Enter 确认   Esc 取消"
	}
	line(5, "│ "+ui.Pad(errText, w-4)+" │")
	line(6, "│ "+ui.Pad("", w-4)+" │")
	line(7, "└"+strings.Repeat("─", w-2)+"┘")
}
