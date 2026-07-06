#!/usr/bin/env bash

_tui_action() {
    case "$1" in
    overview)
        _tui_overview
        ;;
    profiles)
        _tui_profiles
        ;;
    proxies)
        _tui_proxies
        ;;
    tun)
        _tui_tun
        ;;
    logs)
        clashlog
        ;;
    mixin)
        clashmixin -e
        ;;
    webui)
        _webui_show
        ;;
    core)
        _tui_core
        ;;
    settings)
        _tui_settings
        ;;
    esac
}

_tui_profiles_menu() {
    printf 'add\t新增订阅\n'
    printf 'list\t查看订阅\n'
    printf 'use\t切换订阅\n'
    printf 'update\t更新当前订阅\n'
    printf 'delete\t删除订阅\n'
    printf 'log\t订阅日志\n'
}

_tui_profiles() {
    local selected key
    selected=$(_tui_profiles_menu | fzf \
        --height=100% \
        --layout=reverse \
        --border \
        --delimiter=$'\t' \
        --with-nth=2 \
        --prompt='profiles > ' \
        --header='Enter 执行 | q/Esc 返回' \
        --preview='bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && _tui_profiles_block"' \
        --preview-window='right:60%:wrap' \
        --bind='q:abort,esc:abort,ctrl-c:abort') || return
    key=${selected%%$'\t'*}

    case "$key" in
    add)
        clashsub add
        ;;
    list)
        clashsub list
        ;;
    use)
        clashsub use
        ;;
    update)
        clashsub update
        ;;
    delete)
        clashsub del
        ;;
    log)
        clashsub log
        ;;
    esac
}

_tui_tun() {
    local selected key
    selected=$(printf 'status\t查看 TUN 状态\non\t开启 TUN\noff\t关闭 TUN\n' | fzf \
        --height=100% \
        --layout=reverse \
        --border \
        --delimiter=$'\t' \
        --with-nth=2 \
        --prompt='tun > ' \
        --header='Enter 执行 | q/Esc 返回' \
        --preview='bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && tunstatus"' \
        --preview-window='right:60%:wrap' \
        --bind='q:abort,esc:abort,ctrl-c:abort') || return
    key=${selected%%$'\t'*}
    clashtun "$key"
}

_tui_core() {
    local selected key
    selected=$(printf 'status\t查看服务状态\nstart\t启动服务\nstop\t停止服务\nrestart\t重启服务\nupgrade\t升级内核\n' | fzf \
        --height=100% \
        --layout=reverse \
        --border \
        --delimiter=$'\t' \
        --with-nth=2 \
        --prompt='core > ' \
        --header='Enter 执行 | q/Esc 返回' \
        --preview='bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && _tui_core_block"' \
        --preview-window='right:60%:wrap' \
        --bind='q:abort,esc:abort,ctrl-c:abort') || return
    key=${selected%%$'\t'*}

    case "$key" in
    status) clashstatus ;;
    start) clashon -s ;;
    stop) clashoff -s ;;
    restart) service_restart ;;
    upgrade) clashupgrade ;;
    esac
}
