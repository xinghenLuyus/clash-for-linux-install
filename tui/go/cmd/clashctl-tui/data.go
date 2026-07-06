package main

import (
	"os"
	"path/filepath"
	"strings"
)

func (a *app) loadProxyGroups() {
	a.proxyError = ""
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
	if len(groups) == 0 && strings.TrimSpace(out) != "" {
		a.proxyError = oneLine(out)
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
	a.profilesError = ""
	out := a.capture(`"$BIN_YQ" -r '(.use // "" | tostring) as $use | (.profiles // [])[] | [(.id|tostring), (.url // ""), (if ((.id | tostring) == $use) then "true" else "false" end)] | @tsv' "$CLASH_PROFILES_META"`)
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
	if len(profiles) == 0 {
		fallback, err := a.loadProfilesFromMeta()
		if err != nil {
			a.profilesError = err.Error()
		}
		if len(fallback) > 0 {
			profiles = fallback
			a.profilesError = ""
		} else if strings.TrimSpace(out) != "" {
			a.profilesError = oneLine(out)
		}
	}
	a.profiles = profiles
	if a.subSelected >= len(a.profiles) {
		a.subSelected = maxInt(0, len(a.profiles)-1)
	}
}

func (a *app) loadProfilesFromMeta() ([]profile, error) {
	path := filepath.Join(a.home, "resources", "profiles.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	use := ""
	var profiles []profile
	current := profile{}
	inProfile := false
	flush := func() {
		if current.ID == "" && current.URL == "" {
			return
		}
		profiles = append(profiles, current)
		current = profile{}
	}
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "use:") {
			use = yamlScalar(strings.TrimSpace(strings.TrimPrefix(line, "use:")))
			continue
		}
		if strings.HasPrefix(line, "- ") {
			flush()
			inProfile = true
			line = strings.TrimSpace(strings.TrimPrefix(line, "- "))
		}
		if !inProfile || !strings.Contains(line, ":") {
			continue
		}
		parts := strings.SplitN(line, ":", 2)
		key := strings.TrimSpace(parts[0])
		val := yamlScalar(strings.TrimSpace(parts[1]))
		switch key {
		case "id":
			current.ID = val
		case "url":
			current.URL = val
		}
	}
	flush()
	for i := range profiles {
		profiles[i].Current = profiles[i].ID != "" && profiles[i].ID == use
	}
	return profiles, nil
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

func (a *app) currentProfilePathHint() string {
	profiles := a.profiles
	if len(profiles) == 0 {
		if fallback, err := a.loadProfilesFromMeta(); err == nil {
			profiles = fallback
		}
	}
	for _, p := range profiles {
		if p.Current && p.ID != "" {
			return filepath.Join(a.home, "resources", "profiles", p.ID+".yaml")
		}
	}
	if len(profiles) > 0 && profiles[0].ID != "" {
		return filepath.Join(a.home, "resources", "profiles", profiles[0].ID+".yaml")
	}
	return filepath.Join(a.home, "resources", "profiles", "<id>.yaml")
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
