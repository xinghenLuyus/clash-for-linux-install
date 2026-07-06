#!/usr/bin/env bash

clashtui() {
    local tui_bin="${CLASHCTL_HOME}/bin/clashctl-tui"
    [ -x "$tui_bin" ] || {
        _errorcat "未找到 Go TUI：$tui_bin"
        _errorcat "请先构建 tui/go：tui/go/build.sh local"
        return 1
    }
    "$tui_bin"
}
