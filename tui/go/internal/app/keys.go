package app

import (
	"io"
	"os"
)

func readKey(r io.Reader) (string, error) {
	var b [1]byte
	for {
		n, err := r.Read(b[:])
		if err != nil {
			return "", err
		}
		if n == 1 {
			break
		}
	}
	switch b[0] {
	case 3:
		return "ctrl-c", nil
	case 13, 10:
		return "enter", nil
	case 27:
		var seq [2]byte
		n, _ := r.Read(seq[:])
		if n == 2 && seq[0] == '[' {
			switch seq[1] {
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
		return string(b[0]), nil
	}
}

func (a *App) prompt(title string, label string) (string, bool) {
	value := ""
	for {
		a.renderPrompt(title, label, value, "Enter 确认 | Esc 取消")
		b, err := readByte()
		if err != nil {
			return "", false
		}
		switch b {
		case 3, 27:
			return "", false
		case 13, 10:
			return value, true
		case 127, 8:
			rs := []rune(value)
			if len(rs) > 0 {
				value = string(rs[:len(rs)-1])
			}
		default:
			if b >= 32 {
				value += string(b)
			}
		}
	}
}

func (a *App) confirm(title string, label string) (bool, bool) {
	for {
		a.renderPrompt(title, label, "y 是 / n 否", "y 确认 | n/Enter 否 | Esc 取消")
		b, err := readByte()
		if err != nil {
			return false, false
		}
		switch b {
		case 3, 27:
			return false, false
		case 'y', 'Y':
			return true, true
		case 'n', 'N', 13, 10:
			return false, true
		}
	}
}

func readByte() (byte, error) {
	var b [1]byte
	for {
		n, err := os.Stdin.Read(b[:])
		if err != nil {
			return 0, err
		}
		if n == 1 {
			return b[0], nil
		}
	}
}

func (a *App) renderPrompt(title string, label string, value string, hint string) {
	a.render()
	w := a.width * 3 / 5
	if w < 50 {
		w = 50
	}
	if w > a.width-8 {
		w = a.width - 8
	}
	x := (a.width - w) / 2
	y := a.height/2 - 4
	line := func(row int, text string) {
		print("\033[" + itoa(y+row) + ";" + itoa(x+1) + "H")
		print(invert(viewPad(text, w)))
	}
	line(0, "+"+repeat("-", w-2)+"+")
	line(1, "| "+title)
	line(2, "| "+label)
	line(3, "| "+value)
	line(4, "| "+hint)
	line(5, "+"+repeat("-", w-2)+"+")
}

func itoa(v int) string {
	if v == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	n := v
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func repeat(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}

func viewPad(s string, w int) string {
	if len(s) > w {
		return s[:w]
	}
	return s + repeat(" ", w-len(s))
}
