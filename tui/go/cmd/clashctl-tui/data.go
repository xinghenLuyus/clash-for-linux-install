package main

import (
	"path/filepath"
	"strings"
)

func (a *app) loadProxyGroups() {
	out := a.capture("proxy_groups_tsv")
	var groups []proxyGroup
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.Count(line, "\t") < 3 {
			continue
		}
		parts := strings.Split(line, "\t")
		for len(parts) < 4 {
			parts = append(parts, "")
		}
		groups = append(groups, proxyGroup{Name: parts[0], Type: parts[1], Now: parts[2], Count: parts[3]})
	}
	a.proxyGroups = groups
	if a.proxyMember {
		if a.proxyGroupIndex >= len(a.proxyGroups) {
			a.proxyGroupIndex = maxInt(0, len(a.proxyGroups)-1)
		}
	} else if a.subSelected >= len(a.proxyGroups) {
		a.subSelected = maxInt(0, len(a.proxyGroups)-1)
	}
}

func (a *app) loadProxyMembers() {
	group := a.currentProxyGroup()
	if group.Name == "" {
		a.proxyMembers = nil
		return
	}
	out := a.capture("proxy_member_rows " + shellQuote(group.Name))
	var members []proxyMember
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.Count(line, "\t") < 4 {
			continue
		}
		parts := strings.Split(line, "\t")
		for len(parts) < 5 {
			parts = append(parts, "")
		}
		members = append(members, proxyMember{Marker: parts[0], Name: parts[1], Type: parts[2], Provider: parts[3], Caps: parts[4]})
	}
	a.proxyMembers = members
	if a.proxyMember && a.subSelected >= len(a.proxyMembers) {
		a.subSelected = maxInt(0, len(a.proxyMembers)-1)
	}
}

func (a *app) currentProxyGroup() proxyGroup {
	if len(a.proxyGroups) == 0 {
		return proxyGroup{}
	}
	idx := a.subSelected
	if a.proxyMember {
		idx = a.proxyGroupIndex
	}
	if idx < 0 || idx >= len(a.proxyGroups) {
		idx = 0
	}
	return a.proxyGroups[idx]
}

func (a *app) currentProxyMember() proxyMember {
	if len(a.proxyMembers) == 0 {
		return proxyMember{}
	}
	idx := a.subSelected
	if idx < 0 || idx >= len(a.proxyMembers) {
		idx = 0
	}
	return a.proxyMembers[idx]
}

func (a *app) loadProfiles() {
	out := a.capture(`"$BIN_YQ" -r '(.use // "") as $use | (.profiles // [])[] | [(.id|tostring), .url, (if (.id == $use) then "true" else "false" end)] | @tsv' "$CLASH_PROFILES_META"`)
	var profiles []profile
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		if strings.Count(line, "\t") < 2 {
			continue
		}
		parts := strings.Split(line, "\t")
		for len(parts) < 3 {
			parts = append(parts, "")
		}
		profiles = append(profiles, profile{ID: parts[0], URL: parts[1], Current: parts[2] == "true"})
	}
	a.profiles = profiles
	if a.subSelected >= len(a.profiles) {
		a.subSelected = maxInt(0, len(a.profiles)-1)
	}
}

func (a *app) currentProfile() profile {
	if len(a.profiles) == 0 {
		return profile{}
	}
	idx := a.subSelected
	if idx < 0 || idx >= len(a.profiles) {
		idx = 0
	}
	return a.profiles[idx]
}

func (a *app) settingState(item string) string {
	switch item {
	case "开启代理环境", "关闭代理环境":
		if strings.Contains(a.statusCache, "服务:运行中") {
			return "服务运行中"
		}
		return "服务已停止"
	case "开启 TUN", "关闭 TUN":
		if strings.Contains(a.statusCache, "TUN:开启") {
			return "ON"
		}
		return "OFF"
	case "查看 Secret", "修改 Secret":
		return "******"
	case "查看 Mixin":
		return filepath.Join(a.home, "resources", "mixin.yaml")
	case "查看 Runtime":
		return filepath.Join(a.home, "resources", "runtime.yaml")
	default:
		return ""
	}
}
