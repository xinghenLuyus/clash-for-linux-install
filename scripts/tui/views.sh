#!/usr/bin/env bash

_tui_mask_secret() {
    [ -n "$1" ] && printf '******' || printf '未设置'
}

_tui_current_profile() {
    "$BIN_YQ" -r '
      (.use // "" | tostring) as $use |
      if ($use == "" or $use == null) then
        "未启用"
      else
        (.profiles // [] | map(select((.id | tostring) == $use))[0].url // ("[" + ($use | tostring) + "]"))
      end
    ' "$CLASH_PROFILES_META" 2>/dev/null
}

_tui_profiles_block() {
    printf '订阅配置\n\n'
    "$BIN_YQ" -r '
      (.use // "" | tostring) as $use |
      (.profiles // [])[] |
      (if ((.id | tostring) == $use) then "*" else " " end) + " [" + (.id | tostring) + "] " + (.url // "")
    ' "$CLASH_PROFILES_META" 2>/dev/null || printf '暂无订阅\n'
    printf '\n操作：Enter 进入订阅管理，可添加/切换/更新/删除订阅。\n'
}

_tui_proxies_block() {
    printf '策略组与节点\n\n'
    service_is_active >&/dev/null || {
        printf '%s 未运行，请先启动服务。\n' "$CLASHCTL_KERNEL"
        return
    }
    proxy_print_groups 2>/dev/null || printf '未能获取策略组，检查控制器或 secret。\n'
    printf '\n操作：Enter 进入 Proxies 页面；页面内可选择节点、搜索、刷新和测速。\n'
}

_tui_logs_block() {
    printf '最近日志\n\n'
    service_read_log | tail -n 80
}

_tui_mixin_block() {
    printf 'Mixin 配置\n\n'
    printf '文件：%s\n\n' "$CLASH_CONFIG_MIXIN"
    sed -n '1,120p' "$CLASH_CONFIG_MIXIN" 2>/dev/null
}

_tui_webui_block() {
    printf 'Web 控制台\n\n'
    _detect_ext_addr
    printf '内网地址：http://%s:%s/ui\n' "$EXT_IP" "$EXT_PORT"
    printf '公共面板：http://board.zash.run.place\n'
    printf '\n操作：Enter 打印完整 Web 控制台地址。\n'
}

_tui_core_block() {
    printf '内核与服务\n\n'
    _tui_status_block
    printf '\n操作：进入后可启动、停止服务或升级内核。\n'
}

_tui_settings_block() {
    printf '设置\n\n'
    printf 'Proxy Environment  复用 clashon / clashoff\n'
    printf 'TUN Mode           复用 tunon / tunoff\n'
    printf 'Secret             复用 clashsecret\n'
    printf 'Mixin              复用 clashmixin -e\n'
    printf 'Ports/LAN/DNS      后续通过 mixin.yaml + _merge_config_restart 实现\n'
}

_tui_preview() {
    case "$1" in
    overview) _tui_status_block ;;
    profiles) _tui_profiles_block ;;
    proxies) _tui_proxies_block ;;
    tun) tunstatus ;;
    logs) _tui_logs_block ;;
    mixin) _tui_mixin_block ;;
    webui) _tui_webui_block ;;
    core) _tui_core_block ;;
    settings) _tui_settings_block ;;
    *) _tui_status_block ;;
    esac
}
