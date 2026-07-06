package main

import (
	"bufio"
	"errors"
	"os"
	"syscall"
	"time"
)

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

func (a *app) prompt(label string) (string, bool) {
	value := ""
	a.modal = &modal{Title: "输入", Label: label}
	defer func() {
		a.modal = nil
		a.message = ""
	}()
	for {
		a.modal.Value = value
		a.render()
		b := make([]byte, 1)
		_, err := os.Stdin.Read(b)
		if err != nil {
			return "", false
		}
		switch b[0] {
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
			if b[0] >= 32 {
				value += string(b)
			}
		}
	}
}
