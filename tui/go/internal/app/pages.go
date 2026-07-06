package app

import "strings"

func buildPages() []Page {
	return []Page{
		{
			Key:     "overview",
			Title:   "总览",
			Summary: "服务状态与基本信息",
			Preview: func(a *App) string { return a.cli.Clashctl("status").Text() },
			Actions: []Action{
				{"s", "刷新状态", func(a *App) string { return a.cli.Clashctl("status").Text() }},
			},
		},
		{
			Key:     "profiles",
			Title:   "订阅",
			Summary: "新增、使用、更新、删除订阅",
			Preview: func(a *App) string { return a.cli.Clashctl("sub", "list").Text() },
			Actions: []Action{
				{"a", "新增订阅", actionSubAdd},
				{"s", "使用订阅", actionSubUse},
				{"u", "更新订阅", actionSubUpdate},
				{"d", "删除订阅", actionSubDelete},
				{"l", "订阅日志", func(a *App) string { return a.cli.Clashctl("sub", "log").Text() }},
			},
		},
		{
			Key:     "proxies",
			Title:   "代理",
			Summary: "查看策略组、查看节点、切换节点",
			Preview: func(a *App) string { return a.cli.Clashctl("node", "list").Text() },
			Actions: []Action{
				{"g", "查看策略组", func(a *App) string { return a.cli.Clashctl("node", "list").Text() }},
				{"n", "查看组内节点", actionNodeMembers},
				{"s", "切换节点", actionNodeUse},
			},
		},
		{
			Key:     "settings",
			Title:   "设置",
			Summary: "代理环境、服务、TUN、Secret",
			Preview: func(a *App) string { return settingsPreview(a) },
			Actions: []Action{
				{"o", "开启代理", func(a *App) string { return a.cli.Clashctl("on").Text() }},
				{"f", "关闭代理", func(a *App) string { return a.cli.Clashctl("off").Text() }},
				{"t", "TUN 状态", func(a *App) string { return a.cli.Clashctl("tun").Text() }},
				{"1", "开启 TUN", func(a *App) string { return a.cli.Clashctl("tun", "on").Text() }},
				{"0", "关闭 TUN", func(a *App) string { return a.cli.Clashctl("tun", "off").Text() }},
				{"x", "查看 Secret", func(a *App) string { return a.cli.Clashctl("secret").Text() }},
			},
		},
		{
			Key:     "logs",
			Title:   "日志",
			Summary: "查看内核日志",
			Preview: func(a *App) string { return a.cli.Clashctl("log", "-n", "120").Text() },
			Actions: []Action{
				{"l", "刷新日志", func(a *App) string { return a.cli.Clashctl("log", "-n", "120").Text() }},
			},
		},
		{
			Key:     "web",
			Title:   "Web 面板",
			Summary: "显示原始 Web 控制台地址",
			Preview: func(a *App) string { return a.cli.Clashctl("ui").Text() },
			Actions: []Action{
				{"w", "显示 Web 面板", func(a *App) string { return a.cli.Clashctl("ui").Text() }},
			},
		},
	}
}

func actionSubAdd(a *App) string {
	url, ok := a.prompt("新增订阅", "订阅链接")
	if !ok || strings.TrimSpace(url) == "" {
		return "已取消。"
	}
	use, ok := a.confirm("新增订阅", "添加后立即使用？")
	if !ok {
		return "已取消。"
	}
	if use {
		return a.cli.Clashctl("sub", "add", "--use", strings.TrimSpace(url)).Text()
	}
	return a.cli.Clashctl("sub", "add", strings.TrimSpace(url)).Text()
}

func actionSubUse(a *App) string {
	id, ok := a.prompt("使用订阅", "订阅 ID")
	if !ok || strings.TrimSpace(id) == "" {
		return "已取消。"
	}
	return a.cli.Clashctl("sub", "use", strings.TrimSpace(id)).Text()
}

func actionSubUpdate(a *App) string {
	id, ok := a.prompt("更新订阅", "订阅 ID，留空更新当前订阅")
	if !ok {
		return "已取消。"
	}
	if strings.TrimSpace(id) == "" {
		return a.cli.Clashctl("sub", "update").Text()
	}
	return a.cli.Clashctl("sub", "update", strings.TrimSpace(id)).Text()
}

func actionSubDelete(a *App) string {
	id, ok := a.prompt("删除订阅", "订阅 ID")
	if !ok || strings.TrimSpace(id) == "" {
		return "已取消。"
	}
	return a.cli.Clashctl("sub", "del", strings.TrimSpace(id)).Text()
}

func actionNodeMembers(a *App) string {
	group, ok := a.prompt("查看节点", "策略组名称")
	if !ok || strings.TrimSpace(group) == "" {
		return "已取消。"
	}
	return a.cli.Clashctl("node", "list", strings.TrimSpace(group)).Text()
}

func actionNodeUse(a *App) string {
	group, ok := a.prompt("切换节点", "策略组名称")
	if !ok || strings.TrimSpace(group) == "" {
		return "已取消。"
	}
	node, ok := a.prompt("切换节点", "节点名称")
	if !ok || strings.TrimSpace(node) == "" {
		return "已取消。"
	}
	return a.cli.Clashctl("node", "use", strings.TrimSpace(group), strings.TrimSpace(node)).Text()
}

func settingsPreview(a *App) string {
	parts := []string{
		a.cli.Clashctl("status").Text(),
		a.cli.Clashctl("tun").Text(),
	}
	return strings.Join(parts, "\n\n")
}
