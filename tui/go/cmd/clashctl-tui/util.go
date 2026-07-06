package main

import "strings"

func shellQuote(s string) string {
	if s == "" {
		return "''"
	}
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func emptyDash(s string) string {
	if strings.TrimSpace(s) == "" {
		return "-"
	}
	return s
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

func yamlScalar(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.Index(s, " #"); i >= 0 {
		s = strings.TrimSpace(s[:i])
	}
	if len(s) >= 2 {
		if (s[0] == '"' && s[len(s)-1] == '"') || (s[0] == '\'' && s[len(s)-1] == '\'') {
			return s[1 : len(s)-1]
		}
	}
	return s
}
