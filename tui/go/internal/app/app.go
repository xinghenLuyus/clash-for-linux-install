package app

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"

	"clashctl-tui/internal/cli"
)

func New() *App {
	a := &App{cli: cli.NewRunner()}
	a.pages = buildPages()
	a.output = "欢迎使用 clashctl TUI。\n\n本界面只调用原始 clashctl CLI，不重新实现任何订阅、节点或服务逻辑。"
	return a
}

func (a *App) Run() error {
	a.resize()
	if err := enterAltScreen(); err != nil {
		return err
	}
	defer leaveAltScreen()
	restore, err := rawMode()
	if err != nil {
		return err
	}
	defer restore()

	a.refresh()
	a.render()
	for {
		key, err := readKey(os.Stdin)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}
		switch key {
		case "q", "ctrl-c":
			return nil
		case "esc", "left":
			a.focused = false
			a.message = ""
		case "up":
			if !a.focused && a.selected > 0 {
				a.selected--
				a.refresh()
			}
		case "down":
			if !a.focused && a.selected < len(a.pages)-1 {
				a.selected++
				a.refresh()
			}
		case "right", "enter":
			a.focused = true
			a.message = "页面已聚焦，按底部快捷键执行动作"
		case "r":
			a.refresh()
			a.message = "已刷新"
		default:
			if a.focused {
				a.runAction(key)
			}
		}
		a.resize()
		a.render()
	}
}

func (a *App) refresh() {
	p := a.pages[a.selected]
	if p.Preview != nil {
		a.output = p.Preview(a)
	}
}

func (a *App) runAction(key string) {
	p := a.pages[a.selected]
	for _, action := range p.Actions {
		if action.Key == key {
			a.message = "正在执行：" + action.Label
			a.render()
			a.output = action.Run(a)
			a.message = "已执行：" + action.Label
			return
		}
	}
}

func enterAltScreen() error {
	fmt.Print("\033[?1049h\033[?25l")
	return nil
}

func leaveAltScreen() {
	fmt.Print("\033[?25h\033[?1049l")
}

func rawMode() (func(), error) {
	cmd := exec.Command("stty", "-g")
	cmd.Stdin = os.Stdin
	out, err := cmd.Output()
	if err != nil {
		return func() {}, err
	}
	state := string(out)
	raw := exec.Command("stty", "raw", "-echo", "min", "0", "time", "1")
	raw.Stdin = os.Stdin
	if err := raw.Run(); err != nil {
		return func() {}, err
	}
	return func() {
		restore := exec.Command("stty", state)
		restore.Stdin = os.Stdin
		_ = restore.Run()
	}, nil
}
