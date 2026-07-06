#!/usr/bin/env bash

for tui_file in "$CLASHCTL_HOME"/scripts/tui/*.sh; do
    [ -f "$tui_file" ] || continue
    . "$tui_file"
done

_tui_bin() {
    printf '%s/bin/clashctl-tui' "$CLASHCTL_HOME"
}

_tui_can_run_go() {
    local tui_bin
    tui_bin=$(_tui_bin)
    [ -x "$tui_bin" ] && "$tui_bin" --probe >/dev/null 2>&1
}

_tui_help() {
    cat <<EOF

clashctl tui - 全屏 TUI 管理界面

Usage:
  clashctl tui
  clashui

Keys:
  ↑↓/←→   切换页面
  Enter   刷新当前页面
  r       刷新
  q/Esc   退出

EOF
}

_tui_install_prompt() {
    export CLASHCTL_HOME

    _tui_can_run_go || {
        _okcat '💡' '未检测到可运行的 Go TUI，保留原 CLI/Web UI 使用方式'
        _webui_show
        return 1
    }

    [ -t 0 ] || {
        _webui_show
        return 1
    }

    printf '%s' "$(_okcat '✨' '检测到 Go TUI，是否现在进入 clashctl TUI？[y/N] ')"
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

clashtui() {
    case "${1:-}" in
    -h | --help)
        _tui_help
        ;;
    *)
        local tui_bin
        tui_bin=$(_tui_bin)
        _tui_can_run_go || {
            _errorcat "未检测到可运行的 Go TUI：$tui_bin"
            _errorcat "请先运行 tui/go/build-tui.sh 构建，或安装包含 Go TUI 的版本。"
            return 1
        }
        "$tui_bin" "$@"
        ;;
    esac
}
