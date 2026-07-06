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
	items  []string
}

type app struct {
	home         string
	selected     int
	focused      bool
	subSelected  int
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
			{"overview", "总览", "运行状态、配置偏差、关键摘要", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"运行状态", "配置偏差", "关键路径"}},
			{"profiles", "订阅", "订阅列表、启用配置、更新记录", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"订阅列表", "当前订阅", "订阅操作"}},
			{"proxies", "代理", "策略组、节点、延迟与切换", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"策略组", "节点列表", "延迟测试"}},
			{"logs", "日志", "内核日志快速预览", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"最近日志", "日志文件", "订阅日志"}},
			{"settings", "设置", "代理环境、TUN、Secret、Mixin", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"代理环境", "TUN 模式", "Secret 与 Mixin"}},
			{"core", "内核", "服务状态、启动停止、升级", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"服务状态", "服务操作", "内核升级"}},
			{"webui", "Web 面板", "控制台地址与访问方式", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"访问地址", "Secret", "说明"}},
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
			if a.focused && key == "esc" {
				a.leavePage()
				break
			}
			return nil
		case "up":
			if a.focused {
				if a.subSelected > 0 {
					a.subSelected--
					a.message = ""
					a.refresh(false)
				}
			} else if a.selected > 0 {
				a.selected--
				a.message = ""
				a.subSelected = 0
				a.refresh(false)
			}
		case "down":
			if a.focused {
				if a.subSelected < len(a.currentPage().items)-1 {
					a.subSelected++
					a.message = ""
					a.refresh(false)
				}
			} else if a.selected < len(a.pages)-1 {
				a.selected++
				a.message = ""
				a.subSelected = 0
				a.refresh(false)
			}
		case "left":
			a.leavePage()
		case "right":
			a.enterPage()
		case "r":
			a.refresh(true)
			a.message = "已刷新"
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
			if a.focused && row == a.selected {
				left += " >"
			}
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
	footer := a.footer()
	if a.message != "" {
		footer += " | " + a.message
	}
	b.WriteString(invert(pad(footer, a.width)))
	b.WriteString("\033[K")
	fmt.Print(b.String())
}

func (a *app) refresh(includeStatus bool) {
	if includeStatus || a.statusCache == "" {
		a.statusCache = fmt.Sprintf(" clashctl TUI | 内核:%s | %s ", a.kernel(), localizeStatus(oneLine(a.capture("_tui_status_line"))))
	}
	a.contentCache = a.pageContent(a.pages[a.selected].key)
}

func (a *app) currentPage() page {
	return a.pages[a.selected]
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
	if a.focused {
		return a.subPageContent(key)
	}
	switch key {
	case "overview":
		return a.capture("_tui_status_block")
	case "profiles":
		return a.capture("_tui_profiles_block")
	case "proxies":
		return a.capture("_tui_proxies_block")
	case "logs":
		return a.logsPreview()
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
	if !a.focused {
		a.enterPage()
		return
	}
	a.refresh(true)
	a.message = fmt.Sprintf("已刷新：%s", a.currentItem())
}

func (a *app) enterPage() {
	if a.focused {
		a.refresh(true)
		return
	}
	a.focused = true
	a.subSelected = 0
	a.message = ""
	a.refresh(true)
}

func (a *app) leavePage() {
	if !a.focused {
		return
	}
	a.focused = false
	a.message = ""
	a.refresh(false)
}

func (a *app) footer() string {
	if a.focused {
		return "←/Esc 返回 | ↑↓ 选择子项 | Enter 刷新当前子页 | r 刷新 | q 退出"
	}
	return a.currentPage().footer
}

func (a *app) currentItem() string {
	items := a.currentPage().items
	if len(items) == 0 || a.subSelected >= len(items) {
		return ""
	}
	return items[a.subSelected]
}

func (a *app) subPageContent(key string) string {
	var b strings.Builder
	p := a.currentPage()
	b.WriteString(p.title)
	b.WriteString(" / ")
	b.WriteString(a.currentItem())
	b.WriteString("\n\n")
	for i, item := range p.items {
		prefix := "  "
		if i == a.subSelected {
			prefix = "> "
		}
		b.WriteString(prefix)
		b.WriteString(item)
		b.WriteString("\n")
	}
	b.WriteString("\n")

	switch key {
	case "overview":
		return b.String() + a.capture("_tui_status_block")
	case "profiles":
		return b.String() + a.capture("_tui_profiles_block")
	case "proxies":
		return b.String() + a.capture("_tui_proxies_block")
	case "logs":
		return b.String() + a.logsPreview()
	case "settings":
		return b.String() + a.capture("_tui_settings_block")
	case "core":
		return b.String() + a.capture("_tui_core_block")
	case "webui":
		return b.String() + a.capture("_tui_webui_block")
	default:
		return b.String()
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

func (a *app) logsPreview() string {
	logPath := a.logPath()
	if logPath == "" {
		return "最近日志\n\n未找到日志路径。"
	}
	lines, err := readTailLines(logPath, 128*1024, 160)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Sprintf("最近日志\n\n日志文件：%s\n暂无日志文件。服务启动后会自动生成。", logPath)
		}
		return fmt.Sprintf("最近日志\n\n日志文件：%s\n读取失败：%v", logPath, err)
	}
	if len(lines) == 0 {
		return fmt.Sprintf("最近日志\n\n日志文件：%s\n暂无日志。", logPath)
	}
	return fmt.Sprintf("最近日志\n日志文件：%s\n\n%s", logPath, strings.Join(lines, "\n"))
}

func (a *app) logPath() string {
	kernel := a.kernel()
	resources := filepath.Join(a.home, "resources")
	candidates := []string{
		filepath.Join(resources, kernel+".log"),
		filepath.Join("/var/log", kernel+".log"),
	}
	for _, p := range candidates {
		if st, err := os.Stat(p); err == nil && !st.IsDir() {
			return p
		}
	}
	return candidates[0]
}

func readTailLines(path string, maxBytes int64, maxLines int) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := st.Size()
	start := int64(0)
	if size > maxBytes {
		start = size - maxBytes
	}
	if _, err := f.Seek(start, 0); err != nil {
		return nil, err
	}
	buf := make([]byte, size-start)
	n, err := f.Read(buf)
	if err != nil && n == 0 {
		return nil, err
	}
	text := strings.TrimRight(string(buf[:n]), "\n")
	if text == "" {
		return nil, nil
	}
	lines := strings.Split(text, "\n")
	if start > 0 && len(lines) > 0 {
		lines = lines[1:]
	}
	if len(lines) > maxLines {
		lines = lines[len(lines)-maxLines:]
	}
	return lines, nil
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

func localizeStatus(s string) string {
	replacer := strings.NewReplacer(
		"Service:", "服务:",
		"API:", "API:",
		"TUN:", "TUN:",
		"Mode:", "模式:",
		"running", "运行中",
		"stopped", "已停止",
		"ok", "正常",
		"down", "异常",
		"on", "开启",
		"off", "关闭",
	)
	return replacer.Replace(s)
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
