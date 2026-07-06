#!/usr/bin/env bash

. "$CLASHCTL_HOME"/.env

for lib_file in "$CLASHCTL_HOME"/scripts/lib/*.sh; do
    . "$lib_file"
done

for cmd_file in "$CLASHCTL_HOME"/scripts/cmd/*.sh; do
    case "$cmd_file" in *clashctl.*) continue ;; esac
    . "$cmd_file"
done

clashctl() {
    local sub_cmd
    sub_cmd=${1:-help}
    shift

    case $sub_cmd in
    -h | --help | help) sub_cmd=help ;;
    esac

    [ "$sub_cmd" = webui ] && {
        _webui_show "$@"
        return
    }

    local target="clash${sub_cmd}"
    declare -F "$target" >&/dev/null || {
        _failcat "Unknown subcommand: $target"
        _failcat "Use 'clashctl help' for usage information."
        return
    }
    CLASHCTL_DISPATCH_SUB_CMD=$sub_cmd "$target" "$@"
}
