#!/usr/bin/env bash

_tui_settings_menu() {
    printf 'proxy-on\tProxy Environment On\n'
    printf 'proxy-off\tProxy Environment Off\n'
    printf 'tun-on\tTUN Mode On\n'
    printf 'tun-off\tTUN Mode Off\n'
    printf 'secret\tWeb Secret\n'
    printf 'mixin\tEdit Mixin\n'
    printf 'runtime\tRuntime Config\n'
    printf 'webui\tWeb UI\n'
    printf 'ports\tPorts (planned)\n'
    printf 'lan\tAllow LAN (planned)\n'
    printf 'dns\tDNS (planned)\n'
}

_tui_settings() {
    local selected key
    selected=$(_tui_settings_menu | fzf \
        --height=100% \
        --layout=reverse \
        --border \
        --delimiter=$'\t' \
        --with-nth=2 \
        --prompt='settings > ' \
        --header='Enter 执行 | q/Esc 返回' \
        --preview='bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && _tui_settings_preview \"$1\"" -- {1}' \
        --preview-window='right:60%:wrap' \
        --bind='q:abort,esc:abort,ctrl-c:abort') || return
    key=${selected%%$'\t'*}

    case "$key" in
    proxy-on)
        clashon
        ;;
    proxy-off)
        clashoff
        ;;
    tun-on)
        tunon
        ;;
    tun-off)
        tunoff
        ;;
    secret)
        clashsecret
        ;;
    mixin)
        clashmixin -e
        ;;
    runtime)
        clashmixin -r
        ;;
    webui)
        _webui_show
        ;;
    ports | lan | dns)
        _okcat "该设置将通过 mixin.yaml + _merge_config_restart 实现，当前版本先保留入口。"
        ;;
    esac
}

_tui_settings_preview() {
    case "$1" in
    proxy-on | proxy-off)
        printf 'Proxy Environment\n\n'
        printf '复用原 CLI 行为：\n'
        printf '  On  -> clashon\n'
        printf '  Off -> clashoff\n\n'
        printf '这不是 OS 级系统代理；它与当前 clashctl CLI 语义保持一致。\n'
        ;;
    tun-on | tun-off)
        printf 'TUN Mode\n\n'
        tunstatus
        printf '\n复用原 CLI 行为：tunon / tunoff\n'
        ;;
    secret)
        printf 'Web Secret\n\n'
        clashsecret
        ;;
    mixin)
        _tui_mixin_block
        ;;
    runtime)
        printf 'Runtime Config\n\n'
        sed -n '1,120p' "$CLASH_CONFIG_RUNTIME" 2>/dev/null
        ;;
    webui)
        _tui_webui_block
        ;;
    ports)
        printf 'Ports\n\n后续通过修改 mixin.yaml 中 mixed-port / port / socks-port 后 _merge_config_restart 生效。\n'
        ;;
    lan)
        printf 'Allow LAN\n\n后续通过修改 mixin.yaml 中 allow-lan / bind-address 后 _merge_config_restart 生效。\n'
        ;;
    dns)
        printf 'DNS\n\n后续通过修改 mixin.yaml 中 dns 配置后 _merge_config_restart 生效。\n'
        ;;
    *)
        _tui_settings_block
        ;;
    esac
}
