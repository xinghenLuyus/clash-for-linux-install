#!/usr/bin/env bash

_tui_service_label() {
    service_is_active >&/dev/null && printf 'running' || printf 'stopped'
}

_tui_api_label() {
    api_available >&/dev/null && printf 'ok' || printf 'down'
}

_tui_tun_label() {
    tunstatus >&/dev/null && printf 'on' || printf 'off'
}

_tui_mode_label() {
    local live_json value
    service_is_active >&/dev/null && live_json=$(api_configs 2>/dev/null)
    if [ -n "$live_json" ]; then
        value=$("$BIN_YQ" -p json -r '.mode // ""' <<<"$live_json" 2>/dev/null)
    fi
    [ -z "$value" ] && value=$("$BIN_YQ" -r '.mode // "rule"' "$CLASH_CONFIG_RUNTIME" 2>/dev/null)
    printf '%s' "${value:-rule}"
}

_tui_title() {
    printf 'clashctl TUI | %s | Service:%s | API:%s | TUN:%s | Mode:%s' \
        "${CLASHCTL_KERNEL:-core}" \
        "$(_tui_service_label)" \
        "$(_tui_api_label)" \
        "$(_tui_tun_label)" \
        "$(_tui_mode_label)"
}

_tui_page_footer() {
    case "$1" in
    overview)
        printf '↑↓ Menu | Enter Open Overview | Ctrl-R Refresh | q Quit'
        ;;
    profiles)
        printf '↑↓ Menu | Enter Open Profiles | inside: add/use/update/delete/log'
        ;;
    proxies)
        printf '↑↓ Menu | Enter Open Proxies | inside: select/search/delay/refresh'
        ;;
    logs)
        printf '↑↓ Menu | Enter Open Logs | Ctrl-R Refresh | q Quit'
        ;;
    settings)
        printf '↑↓ Menu | Enter Open Settings | inside: Proxy/TUN/Secret/Mixin'
        ;;
    core)
        printf '↑↓ Menu | Enter Open Core | inside: start/stop/restart/upgrade'
        ;;
    webui)
        printf '↑↓ Menu | Enter Show Web UI | q Quit'
        ;;
    *)
        printf '↑↓ Menu | Enter Open | Ctrl-R Refresh | q Quit'
        ;;
    esac
}

_tui_route_rows() {
    printf 'overview\tOverview\t总览、运行状态、配置偏差\n'
    printf 'profiles\tProfiles\t订阅配置管理\n'
    printf 'proxies\tProxies\t策略组、节点切换、测速\n'
    printf 'logs\tLogs\t日志查看\n'
    printf 'settings\tSettings\tProxy Environment、TUN、Secret、Mixin\n'
    printf 'core\tCore\t服务与内核\n'
    printf 'webui\tWeb UI\tWeb 控制台地址\n'
}

_tui_route_preview() {
    local page=$1
    case "$page" in
    overview) _tui_status_block ;;
    profiles) _tui_profiles_block ;;
    proxies) _tui_proxies_block ;;
    logs) _tui_logs_block ;;
    settings) _tui_settings_block ;;
    core) _tui_core_block ;;
    webui) _tui_webui_block ;;
    *) _tui_status_block ;;
    esac
}

_tui_route_preview_with_footer() {
    local page=$1
    _tui_route_preview "$page"
    printf '\n\n%s\n' "$(_tui_page_footer "$page")"
}

_tui_frame_select_page() {
    _tui_route_rows | fzf \
        --height=100% \
        --layout=reverse \
        --border \
        --ansi \
        --cycle \
        --delimiter=$'\t' \
        --with-nth='2,3' \
        --prompt='page > ' \
        --border-label="$(_tui_title)" \
        --header='↑↓ Select | Enter Open | Ctrl-R Refresh | q/Esc Quit' \
        --preview='bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && _tui_route_preview_with_footer \"$1\"" -- {1}' \
        --preview-window='right:66%:wrap' \
        --bind='q:abort,esc:abort,ctrl-c:abort,ctrl-r:reload(bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && _tui_route_rows")'
}
