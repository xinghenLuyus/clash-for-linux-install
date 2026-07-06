package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"golang.org/x/term"
)

type page struct {
	key    string
	title  string
	desc   string
	footer string
}

type app struct {
	home         string
	selected     int
	width        int
	height       int
	lastWidth    int
	lastHeight   int
	message      string
	statusCache  string
	contentCache string
	pages        []page
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--probe" {
		return
	}

	a := newApp()
	if err := a.run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newApp() *app {
	home := os.Getenv("CLASHCTL_HOME")
	if home == "" {
		if wd, err := os.Getwd(); err == nil {
			home = wd
		}
	}
	return &app{
		home: home,
		pages: []page{
			{"overview", "Overview", "live status, drift, recent logs", "Enter refresh | arrows switch | r refresh | q quit"},
			{"profiles", "Profiles", "subscriptions and active profile", "Enter actions | arrows switch | r refresh | q quit"},
			{"proxies", "Proxies", "proxy groups, nodes, delay and select", "Enter node actions | arrows switch | r refresh | q quit"},
			{"logs", "Logs", "recent kernel logs", "Enter tail log | arrows switch | r refresh | q quit"},
			{"settings", "Settings", "proxy env, TUN, secret, mixin", "Enter actions | arrows switch | r refresh | q quit"},
			{"core", "Core", "service, status, upgrade", "Enter actions | arrows switch | r refresh | q quit"},
			{"webui", "Web UI", "dashboard addresses", "Enter show addresses | arrows switch | r refresh | q quit"},
		},
	}
}

func (a *app) run() error {
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to enter raw terminal mode: %w", err)
	}
	defer term.Restore(int(os.Stdin.Fd()), oldState)
	fmt.Print("\033[?1049h\033[?25l\033[?7l")
	defer fmt.Print("\033[?7h\033[?25h\033[?1049l")

	a.resize()
	a.refresh(true)
	a.render()
	reader := bufio.NewReader(os.Stdin)
	for {
		key, err := readKey(reader)
		if err != nil {
			return err
		}
		switch key {
		case "q", "ctrl-c", "esc":
			return nil
		case "up":
			if a.selected > 0 {
				a.selected--
				a.message = ""
				a.refresh(false)
			}
		case "down":
			if a.selected < len(a.pages)-1 {
				a.selected++
				a.message = ""
				a.refresh(false)
			}
		case "left":
			if a.selected > 0 {
				a.selected--
				a.message = ""
				a.refresh(false)
			}
		case "right":
			if a.selected < len(a.pages)-1 {
				a.selected++
				a.message = ""
				a.refresh(false)
			}
		case "r":
			a.refresh(true)
			a.message = "refreshed"
		case "enter":
			a.handleEnter()
		}
		a.resize()
		a.render()
	}
}

func readKey(r *bufio.Reader) (string, error) {
	b, err := r.ReadByte()
	if err != nil {
		return "", err
	}
	switch b {
	case 3:
		return "ctrl-c", nil
	case 13, 10:
		return "enter", nil
	case 27:
		next := readPendingEscBytes(r)
		if len(next) < 2 {
			return "esc", nil
		}
		if next[0] == '[' {
			dir := next[1]
			switch dir {
			case 'A':
				return "up", nil
			case 'B':
				return "down", nil
			case 'C':
				return "right", nil
			case 'D':
				return "left", nil
			}
		}
		return "esc", nil
	default:
		return string(b), nil
	}
}

func readPendingEscBytes(r *bufio.Reader) []byte {
	fd := int(os.Stdin.Fd())
	_ = syscall.SetNonblock(fd, true)
	defer syscall.SetNonblock(fd, false)

	time.Sleep(25 * time.Millisecond)
	var out []byte
	for len(out) < 2 {
		b, err := r.ReadByte()
		if err != nil {
			if errors.Is(err, syscall.EAGAIN) || errors.Is(err, syscall.EWOULDBLOCK) {
				break
			}
			break
		}
		out = append(out, b)
	}
	return out
}

func (a *app) resize() {
	w, h, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		a.width, a.height = 100, 30
		return
	}
	if w < 80 {
		w = 80
	}
	if h < 24 {
		h = 24
	}
	a.width, a.height = w, h
}

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
	b.WriteString(invert(pad(top, a.width)))
	b.WriteString("\033[K\r\n")
	leftW := 24
	rightW := a.width - leftW - 3
	bodyH := a.height - 3
	content := wrapLines(a.contentCache, rightW)
	for row := 0; row < bodyH; row++ {
		left := ""
		if row < len(a.pages) {
			p := a.pages[row]
			prefix := "  "
			if row == a.selected {
				prefix = "> "
			}
			left = prefix + p.title
		}
		right := ""
		if row < len(content) {
			right = content[row]
		}
		b.WriteString(pad(left, leftW))
		b.WriteString(" │ ")
		b.WriteString(pad(right, rightW))
		b.WriteString("\033[K\r\n")
	}
	footer := a.pages[a.selected].footer
	if a.message != "" {
		footer += " | " + a.message
	}
	b.WriteString(invert(pad(footer, a.width)))
	b.WriteString("\033[K")
	fmt.Print(b.String())
}

func (a *app) refresh(includeStatus bool) {
	if includeStatus || a.statusCache == "" {
		a.statusCache = fmt.Sprintf(" clashctl TUI | %s | %s ", a.kernel(), oneLine(a.capture("_tui_status_line")))
	}
	a.contentCache = a.pageContent(a.pages[a.selected].key)
}

func (a *app) kernel() string {
	envPath := filepath.Join(a.home, ".env")
	data, err := os.ReadFile(envPath)
	if err != nil {
		return "core"
	}
	for _, line := range strings.Split(string(data), "\n") {
		if strings.HasPrefix(line, "CLASHCTL_KERNEL=") {
			v := strings.TrimPrefix(line, "CLASHCTL_KERNEL=")
			if v != "" {
				return v
			}
		}
	}
	return "core"
}

func (a *app) pageContent(key string) string {
	switch key {
	case "overview":
		return a.capture("_tui_status_block")
	case "profiles":
		return a.capture("_tui_profiles_block")
	case "proxies":
		return a.capture("_tui_proxies_block")
	case "logs":
		return a.capture("_tui_logs_block")
	case "settings":
		return a.capture("_tui_settings_block")
	case "core":
		return a.capture("_tui_core_block")
	case "webui":
		return a.capture("_tui_webui_block")
	default:
		return ""
	}
}

func (a *app) handleEnter() {
	key := a.pages[a.selected].key
	switch key {
	case "overview":
		a.refresh(true)
		a.message = "overview refreshed"
	case "proxies":
		a.refresh(true)
		a.message = "proxies refreshed"
	case "profiles":
		a.refresh(true)
		a.message = "profiles refreshed"
	case "logs":
		a.refresh(true)
		a.message = "logs refreshed"
	case "settings":
		a.refresh(true)
		a.message = "settings refreshed"
	case "core":
		a.refresh(true)
		a.message = "core refreshed"
	case "webui":
		a.refresh(true)
		a.message = "web ui refreshed"
	}
}

func (a *app) capture(fn string) string {
	cmd := exec.Command("bash", "-lc", fmt.Sprintf(". %q/scripts/cmd/clashctl.sh >/dev/null 2>&1; %s 2>&1", a.home, fn))
	cmd.Env = append(os.Environ(), "CLASHCTL_HOME="+a.home)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out
	if err := cmd.Run(); err != nil && out.Len() == 0 {
		return err.Error()
	}
	return out.String()
}

func pad(s string, width int) string {
	s = trimToWidth(s, width)
	w := displayWidth(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func wrapLines(s string, width int) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		line = stripANSI(line)
		if line == "" {
			lines = append(lines, "")
			continue
		}
		for displayWidth(line) > width {
			head, tail := splitByWidth(line, width)
			lines = append(lines, head)
			line = tail
		}
		lines = append(lines, line)
	}
	return lines
}

func oneLine(s string) string {
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return "Service:unknown"
}

func invert(s string) string { return "\033[7m" + s + "\033[0m" }

func stripANSI(s string) string {
	var out []rune
	inEsc := false
	for _, r := range s {
		if inEsc {
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEsc = false
			}
			continue
		}
		if r == 0x1b {
			inEsc = true
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

func trimToWidth(s string, width int) string {
	head, _ := splitByWidth(stripANSI(s), width)
	return head
}

func splitByWidth(s string, width int) (string, string) {
	if width <= 0 {
		return "", s
	}
	var b strings.Builder
	used := 0
	for i, r := range s {
		rw := runeWidth(r)
		if used+rw > width {
			return b.String(), s[i:]
		}
		b.WriteRune(r)
		used += rw
	}
	return b.String(), ""
}

func displayWidth(s string) int {
	width := 0
	for _, r := range stripANSI(s) {
		width += runeWidth(r)
	}
	return width
}

func runeWidth(r rune) int {
	if r == 0 {
		return 0
	}
	if r < 32 || (r >= 0x7f && r < 0xa0) {
		return 0
	}
	if isCombining(r) {
		return 0
	}
	if isWide(r) {
		return 2
	}
	return 1
}

func isCombining(r rune) bool {
	return (r >= 0x0300 && r <= 0x036f) ||
		(r >= 0x1ab0 && r <= 0x1aff) ||
		(r >= 0x1dc0 && r <= 0x1dff) ||
		(r >= 0x20d0 && r <= 0x20ff) ||
		(r >= 0xfe20 && r <= 0xfe2f)
}

func isWide(r rune) bool {
	return (r >= 0x1100 && r <= 0x115f) ||
		r == 0x2329 || r == 0x232a ||
		(r >= 0x2e80 && r <= 0xa4cf) ||
		(r >= 0xac00 && r <= 0xd7a3) ||
		(r >= 0xf900 && r <= 0xfaff) ||
		(r >= 0xfe10 && r <= 0xfe19) ||
		(r >= 0xfe30 && r <= 0xfe6f) ||
		(r >= 0xff00 && r <= 0xff60) ||
		(r >= 0xffe0 && r <= 0xffe6) ||
		(r >= 0x1f300 && r <= 0x1faff) ||
		(r >= 0x20000 && r <= 0x3fffd)
}
