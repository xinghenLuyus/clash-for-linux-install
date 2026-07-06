package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func (a *app) handleEnter() {
	if !a.focused {
		a.enterPage()
		return
	}
	if a.handlePageEnter() {
		return
	}
	a.executeAction()
}

func (a *app) executeAction() {
	key := a.currentPage().key
	item := a.currentItem()
	out := ""
	switch key {
	case "overview":
		out = a.capture("_tui_status_block")
	case "profiles":
		out = a.executeProfileAction(item)
	case "proxies":
		out = a.executeProxyAction(item)
	case "logs":
		out = a.executeLogAction(item)
	case "settings":
		out = a.executeSettingsAction(item)
	case "core":
		out = a.executeCoreAction(item)
	case "webui":
		out = a.executeWebUIAction(item)
	}
	a.actionOutput = strings.TrimSpace(out)
	a.refresh(true)
	a.message = fmt.Sprintf("已执行：%s", item)
}

func (a *app) executeProfileAction(item string) string {
	switch item {
	case "订阅列表":
		return a.capture("clashsub list")
	case "新增订阅":
		url, ok := a.prompt("请输入订阅链接")
		if !ok || strings.TrimSpace(url) == "" {
			return "已取消新增订阅。"
		}
		useAfterAdd, ok := a.confirm("新增订阅", "添加成功后立即使用该订阅？")
		if !ok {
			return "已取消新增订阅。"
		}
		cmd := "clashsub add "
		if useAfterAdd {
			cmd += "--use "
		}
		cmd += shellQuote(strings.TrimSpace(url))
		return a.withBusy("新增订阅", "正在下载、转换并校验订阅，请等待结果返回。", func() string {
			return a.capture(cmd)
		})
	case "切换订阅":
		id, ok := a.prompt("请输入订阅 ID")
		if !ok || strings.TrimSpace(id) == "" {
			return "已取消切换订阅。"
		}
		return a.withBusy("切换订阅", "正在合并配置并重启内核，请等待结果返回。", func() string {
			return a.capture("clashsub use " + shellQuote(strings.TrimSpace(id)))
		})
	case "更新订阅":
		id, ok := a.prompt("请输入订阅 ID，留空更新当前订阅")
		if !ok {
			return "已取消更新订阅。"
		}
		if strings.TrimSpace(id) == "" {
			return a.withBusy("更新订阅", "正在下载、转换并校验订阅，请等待结果返回。", func() string {
				return a.capture("clashsub update")
			})
		}
		return a.withBusy("更新订阅", "正在下载、转换并校验订阅，请等待结果返回。", func() string {
			return a.capture("clashsub update " + shellQuote(strings.TrimSpace(id)))
		})
	case "删除订阅":
		id, ok := a.prompt("请输入要删除的订阅 ID")
		if !ok || strings.TrimSpace(id) == "" {
			return "已取消删除订阅。"
		}
		return a.withBusy("删除订阅", "正在删除订阅并刷新列表，请等待结果返回。", func() string {
			return a.capture("clashsub del " + shellQuote(strings.TrimSpace(id)))
		})
	case "订阅日志":
		return strings.Join(readOptionalTail(filepath.Join(a.home, "resources", "profiles.log"), 64*1024, 120), "\n")
	default:
		return a.capture("_tui_profiles_block")
	}
}

func (a *app) executeProxyAction(item string) string {
	switch item {
	case "策略组列表":
		return a.capture("proxy_print_groups")
	case "查看组节点":
		group, ok := a.prompt("请输入策略组名称")
		if !ok || strings.TrimSpace(group) == "" {
			return "已取消查看节点。"
		}
		return a.capture("proxy_preview_group " + shellQuote(strings.TrimSpace(group)))
	case "切换节点":
		group, ok := a.prompt("请输入策略组名称")
		if !ok || strings.TrimSpace(group) == "" {
			return "已取消切换节点。"
		}
		member, ok := a.prompt("请输入节点名称")
		if !ok || strings.TrimSpace(member) == "" {
			return "已取消切换节点。"
		}
		return a.capture("proxy_apply " + shellQuote(strings.TrimSpace(group)) + " " + shellQuote(strings.TrimSpace(member)))
	case "策略组测速":
		group, ok := a.prompt("请输入策略组名称")
		if !ok || strings.TrimSpace(group) == "" {
			return "已取消测速。"
		}
		return a.capture("proxy_delay_group_rows " + shellQuote(strings.TrimSpace(group)) + " \"$(proxy_default_delay_url)\" \"$(proxy_default_delay_timeout)\" | proxy_print_delay_rows")
	case "节点测速":
		node, ok := a.prompt("请输入节点名称")
		if !ok || strings.TrimSpace(node) == "" {
			return "已取消测速。"
		}
		return a.capture("proxy_delay_one_row " + shellQuote(strings.TrimSpace(node)) + " \"$(proxy_default_delay_url)\" \"$(proxy_default_delay_timeout)\" | proxy_print_delay_rows")
	default:
		return a.capture("_tui_proxies_block")
	}
}

func (a *app) executeLogAction(item string) string {
	switch item {
	case "日志文件":
		return "当前日志文件：\n" + a.logPath()
	case "订阅日志":
		return strings.Join(readOptionalTail(filepath.Join(a.home, "resources", "profiles.log"), 64*1024, 120), "\n")
	default:
		return a.logsPreview()
	}
}

func (a *app) executeSettingsAction(item string) string {
	switch item {
	case "开启代理环境":
		return a.capture("clashon")
	case "关闭代理环境":
		return a.capture("clashoff")
	case "开启 TUN":
		return a.capture("tunon")
	case "关闭 TUN":
		return a.capture("tunoff")
	case "查看 Secret":
		return a.capture("clashsecret")
	case "修改 Secret":
		secret, ok := a.prompt("请输入新的 Web Secret")
		if !ok {
			return "已取消修改 Secret。"
		}
		return a.capture("clashsecret " + shellQuote(secret))
	case "查看 Mixin":
		return readFilePreview(filepath.Join(a.home, "resources", "mixin.yaml"), 80)
	case "查看 Runtime":
		return readFilePreview(filepath.Join(a.home, "resources", "runtime.yaml"), 120)
	default:
		return a.capture("_tui_settings_block")
	}
}

func (a *app) executeCoreAction(item string) string {
	switch item {
	case "服务状态":
		return a.capture("clashstatus")
	case "启动服务":
		return a.capture("clashon -s")
	case "停止服务":
		return a.capture("clashoff -s")
	case "重启服务":
		return a.capture("service_restart")
	case "升级内核":
		return a.capture("clashupgrade")
	default:
		return a.capture("_tui_core_block")
	}
}

func (a *app) executeWebUIAction(item string) string {
	switch item {
	case "查看 Secret":
		return a.capture("clashsecret")
	default:
		return a.capture("_webui_show")
	}
}

func (a *app) handlePageMove(delta int) bool {
	if !a.focused {
		return false
	}
	switch a.currentPage().key {
	case "proxies":
		if a.proxyMember {
			if len(a.proxyMembers) == 0 {
				return true
			}
			a.proxyMemberIndexAdd(delta)
		} else {
			if len(a.proxyGroups) == 0 {
				return true
			}
			next := a.subSelected + delta
			if next >= 0 && next < len(a.proxyGroups) {
				a.subSelected = next
				a.proxyMembers = nil
				a.loadProxyMembers()
			}
		}
		a.message = ""
		a.actionOutput = ""
		a.refresh(false)
		return true
	case "profiles":
		return a.moveWithinCount(len(a.profiles), delta)
	}
	return a.moveWithinCount(len(a.currentPage().items), delta)
}

func (a *app) proxyMemberIndexAdd(delta int) {
	next := a.subSelected + delta
	if next >= 0 && next < len(a.proxyMembers) {
		a.subSelected = next
	}
}

func (a *app) moveWithinCount(count int, delta int) bool {
	if count == 0 {
		a.message = ""
		return true
	}
	next := a.subSelected + delta
	if next < 0 || next >= count {
		return true
	}
	a.subSelected = next
	a.message = ""
	a.actionOutput = ""
	a.refresh(false)
	return true
}

func (a *app) handlePageForward() bool {
	if !a.focused || a.currentPage().key != "proxies" {
		return false
	}
	if !a.proxyMember {
		a.proxyGroupIndex = a.subSelected
		a.proxyMember = true
		a.subSelected = 0
		a.loadProxyMembers()
		a.refresh(false)
	}
	return true
}

func (a *app) handlePageBack() bool {
	if !a.focused || a.currentPage().key != "proxies" || !a.proxyMember {
		return false
	}
	a.proxyMember = false
	a.subSelected = a.proxyGroupIndex
	a.actionOutput = ""
	a.refresh(false)
	return true
}

func (a *app) handlePageEnter() bool {
	if !a.focused {
		return false
	}
	switch a.currentPage().key {
	case "proxies":
		if !a.proxyMember {
			a.handlePageForward()
			return true
		}
		group := a.currentProxyGroup()
		member := a.currentProxyMember()
		if group.Name == "" || member.Name == "" {
			return true
		}
		a.actionOutput = strings.TrimSpace(a.capture("proxy_apply " + shellQuote(group.Name) + " " + shellQuote(member.Name)))
		a.loadProxyGroups()
		a.loadProxyMembers()
		a.refresh(true)
		a.message = "已切换节点"
		return true
	case "profiles":
		p := a.currentProfile()
		if p.ID == "" {
			return true
		}
		a.actionOutput = strings.TrimSpace(a.withBusy("切换订阅", "正在合并配置并重启内核，请等待结果返回。", func() string {
			return a.capture("clashsub use " + shellQuote(p.ID))
		}))
		a.loadProfiles()
		a.refresh(true)
		a.message = "已切换订阅"
		return true
	case "settings":
		a.executeAction()
		return true
	}
	return false
}

func (a *app) handleShortcut(key string) bool {
	switch a.currentPage().key {
	case "profiles":
		switch key {
		case "a":
			a.actionOutput = a.executeProfileAction("新增订阅")
		case "u":
			p := a.currentProfile()
			if p.ID == "" {
				a.actionOutput = "当前没有可更新的订阅。"
			} else {
				a.actionOutput = a.withBusy("更新订阅", "正在下载、转换并校验订阅，请等待结果返回。", func() string {
					return a.capture("clashsub update " + shellQuote(p.ID))
				})
			}
		case "x":
			p := a.currentProfile()
			if p.ID == "" {
				a.actionOutput = "当前没有可删除的订阅。"
			} else {
				a.actionOutput = a.withBusy("删除订阅", "正在删除订阅并刷新列表，请等待结果返回。", func() string {
					return a.capture("clashsub del " + shellQuote(p.ID))
				})
			}
		default:
			return false
		}
		a.loadProfiles()
		a.refresh(true)
		return true
	case "proxies":
		switch key {
		case "d":
			if a.proxyMember {
				member := a.currentProxyMember()
				if member.Name != "" {
					a.actionOutput = a.capture("proxy_delay_one_row " + shellQuote(member.Name) + " \"$(proxy_default_delay_url)\" \"$(proxy_default_delay_timeout)\" | proxy_print_delay_rows")
				}
			}
		case "D":
			group := a.currentProxyGroup()
			if group.Name != "" {
				a.actionOutput = a.capture("proxy_delay_group_rows " + shellQuote(group.Name) + " \"$(proxy_default_delay_url)\" \"$(proxy_default_delay_timeout)\" | proxy_print_delay_rows")
			}
		default:
			return false
		}
		a.refresh(true)
		return true
	}
	return false
}
