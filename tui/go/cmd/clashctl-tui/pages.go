package main

import (
	"fmt"
	"strings"

	"clashctl/internal/ui"
)

func (a *app) pageContent(key string) string {
	switch key {
	case "overview":
		if a.focused {
			return a.subPageContent(key)
		}
		return a.capture("_tui_status_block")
	case "profiles":
		return a.profilesPage()
	case "proxies":
		return a.proxiesPage()
	case "logs":
		if a.focused {
			return a.subPageContent(key)
		}
		return a.logsPreview()
	case "settings":
		return a.settingsPage()
	case "core":
		if a.focused {
			return a.subPageContent(key)
		}
		return a.capture("_tui_core_block")
	case "webui":
		if a.focused {
			return a.subPageContent(key)
		}
		return a.capture("_tui_webui_block")
	default:
		return ""
	}
}

func (a *app) proxiesPage() string {
	if len(a.proxyGroups) == 0 {
		a.loadProxyGroups()
	}
	if len(a.proxyGroups) > 0 && len(a.proxyMembers) == 0 {
		a.loadProxyMembers()
	}
	var b strings.Builder
	b.WriteString("代理\n\n")
	b.WriteString("策略组")
	if a.proxyMember {
		b.WriteString("                    节点\n")
	} else {
		b.WriteString("  <当前焦点>          节点\n")
	}
	max := maxInt(len(a.proxyGroups), len(a.proxyMembers))
	if max == 0 {
		b.WriteString("\n未获取到策略组。\n\n")
		if a.proxyError != "" {
			b.WriteString("读取异常：")
			b.WriteString(a.proxyError)
			b.WriteString("\n\n")
		}
		b.WriteString("可按 r 刷新，或在终端检查：\n")
		b.WriteString("  clashctl node list\n")
		b.WriteString("  clashctl sub list\n\n")
		b.WriteString("配置文件位置：\n")
		b.WriteString("  当前订阅：")
		b.WriteString(a.currentProfilePathHint())
		b.WriteByte('\n')
		b.WriteString("  基础配置：")
		b.WriteString(a.home)
		b.WriteString("/resources/config.yaml\n")
		b.WriteString("  运行配置：")
		b.WriteString(a.home)
		b.WriteString("/resources/runtime.yaml\n")
		b.WriteString("  转换日志：")
		b.WriteString(a.home)
		b.WriteString("/bin/subconverter/latest.log\n")
		return b.String()
	}
	for i := 0; i < max; i++ {
		left := ""
		if i < len(a.proxyGroups) {
			g := a.proxyGroups[i]
			prefix := "  "
			if !a.proxyMember && i == a.subSelected {
				prefix = "> "
			}
			left = fmt.Sprintf("%s%s -> %s", prefix, g.Name, emptyDash(g.Now))
		}
		right := ""
		if i < len(a.proxyMembers) {
			m := a.proxyMembers[i]
			prefix := "  "
			if a.proxyMember && i == a.subSelected {
				prefix = "> "
			}
			mark := " "
			if m.Marker == "*" {
				mark = "*"
			}
			right = fmt.Sprintf("%s%s %s [%s]", prefix, mark, m.Name, emptyDash(m.Type))
		}
		b.WriteString(ui.Pad(left, 34))
		b.WriteString("  ")
		b.WriteString(right)
		b.WriteByte('\n')
	}
	b.WriteString("\n")
	if a.actionOutput != "" {
		b.WriteString("执行结果\n")
		b.WriteString(a.actionOutput)
		b.WriteByte('\n')
	} else {
		b.WriteString("Enter/→ 进入节点列表 | ← 返回策略组 | Enter 切换节点 | d 测速节点 | D 测速策略组\n")
	}
	return b.String()
}

func (a *app) profilesPage() string {
	if len(a.profiles) == 0 {
		a.loadProfiles()
	}
	var b strings.Builder
	b.WriteString("订阅\n\n")
	if len(a.profiles) == 0 {
		b.WriteString("暂无订阅。\n\n")
		if a.profilesError != "" {
			b.WriteString("订阅列表读取异常：")
			b.WriteString(a.profilesError)
			b.WriteString("\n\n")
			b.WriteString("可按 Enter 查看原始订阅配置，或按 r 重新刷新。\n")
		} else {
			b.WriteString("按 a 新增订阅。新增完成后会在这里显示结果。\n")
		}
		if a.actionOutput != "" {
			b.WriteString("\n执行结果\n")
			b.WriteString(a.actionOutput)
			b.WriteByte('\n')
		}
		return b.String()
	}
	for i, p := range a.profiles {
		prefix := "  "
		if i == a.subSelected {
			prefix = "> "
		}
		mark := " "
		if p.Current {
			mark = "*"
		}
		b.WriteString(fmt.Sprintf("%s%s [%s] %s\n", prefix, mark, p.ID, p.URL))
	}
	b.WriteString("\nEnter 使用选中订阅 | a 新增 | u 更新选中 | x 删除选中 | r 刷新\n")
	if a.actionOutput != "" {
		b.WriteString("\n执行结果\n")
		b.WriteString(a.actionOutput)
		b.WriteByte('\n')
	}
	return b.String()
}

func (a *app) settingsPage() string {
	items := a.currentPage().items
	var b strings.Builder
	b.WriteString("设置\n\n")
	for i, item := range items {
		prefix := "  "
		if i == a.subSelected {
			prefix = "> "
		}
		b.WriteString(prefix)
		b.WriteString(item)
		b.WriteString("    ")
		b.WriteString(a.settingState(item))
		b.WriteByte('\n')
	}
	b.WriteString("\nEnter 执行/切换 | ↑↓ 选择 | ← 返回\n")
	if a.actionOutput != "" {
		b.WriteString("\n执行结果\n")
		b.WriteString(a.actionOutput)
		b.WriteByte('\n')
	}
	return b.String()
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
	b.WriteString(a.subPageHint(key))
	b.WriteString("\n\n")

	if a.actionOutput != "" {
		b.WriteString("执行结果\n")
		b.WriteString(a.actionOutput)
		return b.String()
	}

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

func (a *app) subPageHint(key string) string {
	item := a.currentItem()
	switch key + "/" + item {
	case "profiles/新增订阅":
		return "输入订阅链接后添加到 Profiles。"
	case "profiles/切换订阅":
		return "输入订阅 ID 后切换当前配置，并重新合并运行配置。"
	case "profiles/更新订阅":
		return "输入订阅 ID 更新指定订阅；留空更新当前订阅。"
	case "profiles/删除订阅":
		return "输入订阅 ID 删除未启用的订阅。"
	case "proxies/查看组节点":
		return "输入策略组名称，查看该组当前节点和候选节点。"
	case "proxies/切换节点":
		return "依次输入策略组名称和节点名称，直接调用 Clash API 切换。"
	case "proxies/策略组测速":
		return "输入策略组名称，对组内节点测速。"
	case "proxies/节点测速":
		return "输入节点名称，对单个节点测速。"
	case "settings/开启代理环境":
		return "复用 clashon：启动服务并设置当前终端代理环境。"
	case "settings/关闭代理环境":
		return "复用 clashoff：停止服务并清理当前终端代理环境。"
	case "settings/开启 TUN":
		return "复用 tunon：修改 mixin 并重启内核。"
	case "settings/关闭 TUN":
		return "复用 tunoff：关闭 TUN 并重启内核。"
	case "settings/修改 Secret":
		return "输入新的 Web Secret，写入 mixin 并重启生效。"
	case "core/升级内核":
		return "复用 clashupgrade，通过控制器 API 触发内核升级。"
	default:
		return "Enter 执行当前子项；结果会显示在右侧面板内。"
	}
}
