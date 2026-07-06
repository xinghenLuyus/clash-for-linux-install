package app

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"clashctl-tui/internal/view"
)

func (a *App) resize() {
	cmd := exec.Command("stty", "size")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		a.width, a.height = 100, 30
		return
	}
	fields := strings.Fields(string(out))
	if len(fields) != 2 {
		a.width, a.height = 100, 30
		return
	}
	h, _ := strconv.Atoi(fields[0])
	w, _ := strconv.Atoi(fields[1])
	if w < 80 {
		w = 80
	}
	if h < 24 {
		h = 24
	}
	a.width, a.height = w, h
}

func (a *App) render() {
	var b strings.Builder
	b.Grow(a.width * a.height)
	b.WriteString("\033[H\033[2J")
	top := " clashctl TUI | CLI passthrough mode "
	b.WriteString(invert(view.Pad(top, a.width)))
	b.WriteString("\r\n")

	leftW := 22
	rightW := a.width - leftW - 3
	bodyH := a.height - 3
	content := view.Wrap(a.output, rightW, bodyH)

	for row := 0; row < bodyH; row++ {
		left := ""
		if row < len(a.pages) {
			p := a.pages[row]
			prefix := "  "
			if row == a.selected {
				prefix = "> "
			}
			left = prefix + p.Title
			if a.focused && row == a.selected {
				left += " >"
			}
		}
		right := ""
		if row < len(content) {
			right = content[row]
		}
		b.WriteString(view.Pad(left, leftW))
		b.WriteString(" │ ")
		b.WriteString(view.Pad(right, rightW))
		b.WriteString("\r\n")
	}

	footer := a.footer()
	if a.message != "" {
		footer += " | " + a.message
	}
	b.WriteString(invert(view.Pad(footer, a.width)))
	fmt.Print(b.String())
}

func (a *App) footer() string {
	if !a.focused {
		return "↑↓ 选择页面 | →/Enter 进入页面 | r 刷新 | q 退出"
	}
	labels := []string{"←/Esc 返回"}
	for _, action := range a.pages[a.selected].Actions {
		labels = append(labels, action.Key+" "+action.Label)
	}
	labels = append(labels, "r 刷新", "q 退出")
	return strings.Join(labels, " | ")
}

func invert(s string) string {
	return "\033[7m" + s + "\033[0m"
}
