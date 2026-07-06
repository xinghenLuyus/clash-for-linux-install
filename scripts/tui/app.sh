#!/usr/bin/env bash

_tui_app_main() {
    _tui_require_fzf || return
    export CLASHCTL_HOME

    local selected page
    while true; do
        selected=$(_tui_frame_select_page) || break
        page=${selected%%$'\t'*}
        [ -n "$page" ] || continue
        _tui_action "$page"
    done
}
