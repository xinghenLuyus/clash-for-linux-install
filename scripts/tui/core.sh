#!/usr/bin/env bash

_tui_has_fzf() {
    command -v fzf >&/dev/null
}

_tui_require_fzf() {
    _tui_has_fzf && return 0
    _errorcat "未检测到 fzf，无法进入 TUI，请先安装 fzf 后重试"
    return 1
}

_tui_help() {
    cat <<EOF

clashctl tui - 全屏 TUI 管理界面

Usage:
  clashctl tui
  clashui

Keys:
  Enter   执行当前动作
  Ctrl-R  刷新当前界面
  q/Esc   退出

EOF
}

_tui_menu() {
    printf 'overview\t总览\n'
    printf 'profiles\t订阅配置\n'
    printf 'proxies\t策略组与节点\n'
    printf 'tun\tTUN 模式\n'
    printf 'logs\t日志\n'
    printf 'mixin\tMixin 配置\n'
    printf 'webui\tWeb 控制台\n'
    printf 'core\t内核与服务\n'
    printf 'settings\t设置\n'
}

_tui_fzf() {
    fzf \
        --height=100% \
        --layout=reverse \
        --border \
        --ansi \
        --cycle \
        --delimiter=$'\t' \
        --with-nth=2 \
        --prompt='clashctl > ' \
        --header='Enter 执行 | Ctrl-R 刷新 | q/Esc/Ctrl-C 退出' \
        --preview='bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && _tui_preview \"$1\"" -- {1}' \
        --preview-window='right:60%:wrap' \
        --bind='q:abort,esc:abort,ctrl-c:abort,ctrl-r:reload(bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && _tui_menu")'
}

_tui_main() {
    _tui_require_fzf || return
    export CLASHCTL_HOME

    local selected key
    while true; do
        selected=$(_tui_menu | _tui_fzf) || break
        key=${selected%%$'\t'*}
        [ -n "$key" ] || continue
        _tui_action "$key"
        printf '\n'
        printf '按 Enter 返回 TUI，或按 q 后 Enter 退出：'
        local choice
        read -r choice
        [ "$choice" = q ] && break
    done
}

_tui_install_prompt() {
    export CLASHCTL_HOME

    _tui_has_fzf || {
        _okcat '💡' '未检测到 fzf，保留原 CLI/Web UI 使用方式'
        _webui_show
        return 1
    }

    [ -t 0 ] || {
        _webui_show
        return 1
    }

    printf '%s' "$(_okcat '✨' '检测到 fzf，是否现在进入 clashctl TUI？[y/N] ')"
    local answer
    read -r answer
    case "$answer" in
    y | Y | yes | YES)
        clashtui
        return 0
        ;;
    *)
        _webui_show
        return 1
        ;;
    esac
}
