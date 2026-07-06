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

_tui_main() {
    _tui_app_main "$@"
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
