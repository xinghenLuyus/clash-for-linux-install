package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type page struct {
	key    string
	title  string
	desc   string
	footer string
}

type app struct {
	home     string
	selected int
	width    int
	height   int
	message  string
	pages    []page
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
	if err := shell("stty", "raw", "-echo", "min", "1", "time", "0"); err != nil {
		return fmt.Errorf("failed to enter raw terminal mode: %w", err)
	}
	defer shell("stty", "sane")
	fmt.Print("\033[?1049h\033[?25l")
	defer fmt.Print("\033[?25h\033[?1049l")

	a.resize()
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
			}
		case "down":
			if a.selected < len(a.pages)-1 {
				a.selected++
			}
		case "left":
			if a.selected > 0 {
				a.selected--
			}
		case "right":
			if a.selected < len(a.pages)-1 {
				a.selected++
			}
		case "r":
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

func (a *app) render() {
	fmt.Print("\033[H\033[2J")
	top := a.statusLine()
	fmt.Print(invert(pad(top, a.width)) + "\r\n")
	leftW := 24
	rightW := a.width - leftW - 3
	bodyH := a.height - 3
	content := wrapLines(a.pageContent(a.pages[a.selected].key), rightW)
	for row := 0; row < bodyH; row++ {
		left := ""
		if row < len(a.pages) {
			p := a.pages[row]
			prefix := "  "
			if row == a.selected {
				prefix = "> "
			}
			left = prefix + p.title
			if row == a.selected {
				left = cyan(left)
			}
		}
		right := ""
		if row < len(content) {
			right = content[row]
		}
		fmt.Printf("%s │ %s\r\n", pad(left, leftW), pad(right, rightW))
	}
	footer := a.pages[a.selected].footer
	if a.message != "" {
		footer += " | " + a.message
	}
	fmt.Print(invert(pad(footer, a.width)) + "\r\n")
}

func (a *app) statusLine() string {
	return fmt.Sprintf(" clashctl TUI | %s | %s ", a.kernel(), oneLine(a.capture("_tui_status_line")))
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
		a.message = "overview refreshed"
	case "proxies":
		a.message = "proxies refreshed"
	case "profiles":
		a.message = "profiles refreshed"
	case "logs":
		a.message = "logs refreshed"
	case "settings":
		a.message = "settings refreshed"
	case "core":
		a.message = "core refreshed"
	case "webui":
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

func (a *app) runShell(fn string) {
	fmt.Print("\033[?25h\033[?1049l")
	_ = shell("stty", "sane")
	cmd := exec.Command("bash", "-lc", fmt.Sprintf(". %q/scripts/cmd/clashctl.sh; %s", a.home, fn))
	cmd.Env = append(os.Environ(), "CLASHCTL_HOME="+a.home)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	fmt.Print("\nPress Enter to return to clashctl TUI...")
	_, _ = bufio.NewReader(os.Stdin).ReadString('\n')
	_ = shell("stty", "raw", "-echo", "min", "1", "time", "0")
	fmt.Print("\033[?1049h\033[?25l")
	if err != nil {
		a.message = err.Error()
	} else {
		a.message = "done"
	}
}

func shell(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func pad(s string, width int) string {
	rs := []rune(stripANSI(s))
	if len(rs) >= width {
		return string([]rune(s)[:minRuneLen(s, width)])
	}
	return s + strings.Repeat(" ", width-len(rs))
}

func minRuneLen(s string, width int) int {
	rs := []rune(s)
	if len(rs) < width {
		return len(rs)
	}
	return width
}

func wrapLines(s string, width int) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		rs := []rune(line)
		if len(rs) == 0 {
			lines = append(lines, "")
			continue
		}
		for len(rs) > width {
			lines = append(lines, string(rs[:width]))
			rs = rs[width:]
		}
		lines = append(lines, string(rs))
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
func cyan(s string) string   { return "\033[36m" + s + "\033[0m" }

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
