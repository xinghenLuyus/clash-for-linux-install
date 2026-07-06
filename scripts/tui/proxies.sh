#!/usr/bin/env bash

_tui_proxies() {
    service_is_active >&/dev/null || {
        _failcat "$CLASHCTL_KERNEL 未运行，请先启动服务"
        return 1
    }

    local selected group member
    while true; do
        selected=$(proxy_groups_tsv | fzf \
            --height=100% \
            --layout=reverse \
            --border \
            --delimiter=$'\t' \
            --with-nth='1,3,2,4' \
            --prompt='proxies > ' \
            --header='Enter 进入策略组 | Ctrl-D 测速当前组 | Ctrl-R 刷新 | q/Esc 返回' \
            --preview='bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && proxy_preview_group \"$1\"" -- {1}' \
            --preview-window='right:60%:wrap' \
            --bind='q:abort,esc:abort,ctrl-c:abort,ctrl-r:reload(bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && proxy_groups_tsv")' \
            --bind='ctrl-d:execute(bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && printf \"策略组测速：%s\n\n\" \"$1\" && proxy_delay_group_rows \"$1\" \"$(proxy_default_delay_url)\" \"$(proxy_default_delay_timeout)\" | proxy_print_delay_rows && printf \"\n按 Enter 返回...\" && read _" -- {1})') || return
        group=${selected%%$'\t'*}
        [ -n "$group" ] || continue
        member=$(_tui_proxy_members "$group") || continue
        [ -n "$member" ] && proxy_apply "$group" "$member"
        printf '\n按 Enter 返回 Proxies，或 q 后 Enter 返回主菜单：'
        local choice
        read -r choice
        [ "$choice" = q ] && return
    done
}

_tui_proxy_members() {
    local group=$1 selected
    selected=$(proxy_member_rows "$group" | fzf \
        --height=100% \
        --layout=reverse \
        --border \
        --delimiter=$'\t' \
        --with-nth='1,2,3,4,5' \
        --prompt="$group > " \
        --header='Enter 切换节点 | Ctrl-D 测速当前节点 | q/Esc 返回 | / 搜索' \
        --preview='bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && proxy_preview_member \"$1\" \"$2\"" -- "$group" {2}' \
        --preview-window='right:50%:wrap' \
        --bind='q:abort,esc:abort,ctrl-c:abort' \
        --bind='ctrl-d:execute(bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && printf \"节点测速：%s\n\n\" \"$1\" && proxy_delay_one_row \"$1\" \"$(proxy_default_delay_url)\" \"$(proxy_default_delay_timeout)\" | proxy_print_delay_rows && printf \"\n按 Enter 返回...\" && read _" -- {2})') || return
    printf '%s\n' "$(printf '%s' "$selected" | awk -F '\t' '{print $2}')"
}
