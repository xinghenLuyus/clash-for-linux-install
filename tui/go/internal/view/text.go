package view

import (
	"regexp"
	"strings"
	"unicode/utf8"
)

var ansiRE = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func StripANSI(s string) string {
	return ansiRE.ReplaceAllString(s, "")
}

func OneLine(s string) string {
	for _, line := range strings.Split(StripANSI(s), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			return line
		}
	}
	return ""
}

func Truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	clean := StripANSI(s)
	if DisplayWidth(clean) <= width {
		return clean
	}
	rs := []rune(clean)
	for len(rs) > 0 && DisplayWidth(string(rs))+1 > width {
		rs = rs[:len(rs)-1]
	}
	return string(rs) + "…"
}

func Pad(s string, width int) string {
	s = Truncate(s, width)
	pad := width - DisplayWidth(s)
	if pad < 0 {
		pad = 0
	}
	return s + strings.Repeat(" ", pad)
}

func Wrap(s string, width int, maxLines int) []string {
	if width <= 0 {
		return nil
	}
	var out []string
	for _, raw := range strings.Split(StripANSI(s), "\n") {
		line := raw
		if line == "" {
			out = append(out, "")
			continue
		}
		for DisplayWidth(line) > width {
			cut := cutWidth(line, width)
			out = append(out, line[:cut])
			line = line[cut:]
			if maxLines > 0 && len(out) >= maxLines {
				return out
			}
		}
		out = append(out, line)
		if maxLines > 0 && len(out) >= maxLines {
			return out
		}
	}
	return out
}

func cutWidth(s string, width int) int {
	used := 0
	for i, r := range s {
		w := runeWidth(r)
		if used+w > width {
			if i == 0 {
				_, size := utf8.DecodeRuneInString(s)
				return size
			}
			return i
		}
		used += w
	}
	return len(s)
}

func DisplayWidth(s string) int {
	w := 0
	for _, r := range s {
		w += runeWidth(r)
	}
	return w
}

func runeWidth(r rune) int {
	switch {
	case r == 0:
		return 0
	case r < 32:
		return 0
	case r >= 0x1100 && r <= 0x115F:
		return 2
	case r >= 0x2E80 && r <= 0xA4CF:
		return 2
	case r >= 0xAC00 && r <= 0xD7A3:
		return 2
	case r >= 0xF900 && r <= 0xFAFF:
		return 2
	case r >= 0xFE30 && r <= 0xFE4F:
		return 2
	case r >= 0xFF00 && r <= 0xFF60:
		return 2
	case r >= 0xFFE0 && r <= 0xFFE6:
		return 2
	case r >= 0x1F300 && r <= 0x1FAFF:
		return 2
	default:
		return 1
	}
}
