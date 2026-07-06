#!/usr/bin/env bash

clashui() {
    local from_legacy_ui=false
    [ "${CLASHCTL_DISPATCH_SUB_CMD:-}" = ui ] && from_legacy_ui=true

    case "${1:-}" in
    -h | --help)
        if [ "$from_legacy_ui" = true ]; then
            cat <<EOF

原 clashctl ui 已更名为 clashctl webui。

Usage:
  clashctl ui      兼容入口，打开 TUI
  clashctl webui   查看 Web 控制台地址
  clashui          直接打开 TUI

EOF
            return 0
        fi
        clashtui -h
        return 0
        ;;
    esac

    [ "$from_legacy_ui" = true ] && _okcat 'ℹ️ ' '原 clashctl ui 已更名为 clashctl webui。'

    if declare -F clashtui >&/dev/null; then
        clashtui "$@"
        return
    fi

    _errorcat "TUI 未加载，请检查 clashctl 安装是否完整"
    return 1
}
