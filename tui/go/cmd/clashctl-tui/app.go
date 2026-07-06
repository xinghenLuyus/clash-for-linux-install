package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/term"
)

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
			{"profiles", "订阅", "订阅列表、启用配置、更新记录", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"订阅列表", "新增订阅", "切换订阅", "更新订阅", "删除订阅", "订阅日志"}},
			{"proxies", "代理", "策略组、节点、延迟与切换", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"策略组列表", "查看组节点", "切换节点", "策略组测速", "节点测速"}},
			{"logs", "日志", "内核日志快速预览", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"最近日志", "日志文件", "订阅日志"}},
			{"settings", "设置", "代理环境、TUN、Secret、Mixin", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"开启代理环境", "关闭代理环境", "开启 TUN", "关闭 TUN", "查看 Secret", "修改 Secret", "查看 Mixin", "查看 Runtime"}},
			{"core", "内核", "服务状态、启动停止、升级", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"服务状态", "启动服务", "停止服务", "重启服务", "升级内核"}},
			{"webui", "Web 面板", "控制台地址与访问方式", "↑↓ 选择页面 | →/Enter 进入 | r 刷新 | q 退出", []string{"访问地址", "查看 Secret", "刷新地址"}},
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
		case "a", "u", "x", "d", "D":
			if a.focused && a.handleShortcut(key) {
				break
			}
		case "up":
			if a.handlePageMove(-1) {
				break
			}
			a.moveNavigation(-1)
		case "down":
			if a.handlePageMove(1) {
				break
			}
			a.moveNavigation(1)
		case "left":
			if !a.handlePageBack() {
				a.leavePage()
			}
		case "right":
			if !a.handlePageForward() {
				a.enterPage()
			}
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

func (a *app) refresh(includeStatus bool) {
	if includeStatus || a.statusCache == "" {
		a.statusCache = fmt.Sprintf(" clashctl TUI | 内核:%s | %s ", a.kernel(), localizeStatus(oneLine(a.capture("_tui_status_line"))))
	}
	if includeStatus {
		a.refreshPageData()
	}
	a.contentCache = a.pageContent(a.pages[a.selected].key)
}

func (a *app) refreshPageData() {
	switch a.currentPage().key {
	case "proxies":
		a.loadProxyGroups()
	case "profiles":
		a.loadProfiles()
	}
	a.clampSelection()
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

func (a *app) enterPage() {
	if a.focused {
		a.refresh(true)
		return
	}
	a.focused = true
	a.subSelected = 0
	a.proxyMember = false
	a.proxyGroupIndex = 0
	a.message = ""
	a.actionOutput = ""
	a.refresh(true)
}

func (a *app) leavePage() {
	if !a.focused {
		return
	}
	a.focused = false
	a.message = ""
	a.actionOutput = ""
	a.refresh(false)
}

func (a *app) footer() string {
	if a.focused {
		switch a.currentPage().key {
		case "proxies":
			return "← 返回导航/策略组 | → 进入节点 | ↑↓ 选择 | Enter 切换 | d/D 测速 | r 刷新 | q 退出"
		case "profiles":
			return "← 返回导航 | ↑↓ 选择 | Enter 使用 | a 新增 | u 更新 | x 删除 | r 刷新 | q 退出"
		case "settings":
			return "← 返回导航 | ↑↓ 选择 | Enter 执行/切换 | r 刷新 | q 退出"
		default:
			return "←/Esc 返回导航 | ↑↓ 选择 | Enter 执行 | r 刷新 | q 退出"
		}
	}
	return "↑↓ 选择页面 | →/Enter 操作页面 | r 刷新 | q 退出"
}

func (a *app) currentItem() string {
	items := a.currentPage().items
	if len(items) == 0 || a.subSelected >= len(items) {
		return ""
	}
	return items[a.subSelected]
}

func (a *app) moveNavigation(delta int) {
	if a.focused {
		return
	}
	next := a.selected + delta
	if next < 0 || next >= len(a.pages) {
		return
	}
	a.selected = next
	a.message = ""
	a.subSelected = 0
	a.proxyMember = false
	a.proxyGroupIndex = 0
	a.actionOutput = ""
	a.refresh(false)
}

func (a *app) clampSelection() {
	if a.subSelected < 0 {
		a.subSelected = 0
	}
	switch a.currentPage().key {
	case "proxies":
		if a.proxyMember {
			if a.subSelected >= len(a.proxyMembers) {
				a.subSelected = maxInt(0, len(a.proxyMembers)-1)
			}
			return
		}
		if a.subSelected >= len(a.proxyGroups) {
			a.subSelected = maxInt(0, len(a.proxyGroups)-1)
		}
	case "profiles":
		if a.subSelected >= len(a.profiles) {
			a.subSelected = maxInt(0, len(a.profiles)-1)
		}
	default:
		items := a.currentPage().items
		if a.subSelected >= len(items) {
			a.subSelected = maxInt(0, len(items)-1)
		}
	}
}
