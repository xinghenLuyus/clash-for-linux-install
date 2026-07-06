#!/usr/bin/env bash

for tui_file in "$CLASHCTL_HOME"/scripts/tui/*.sh; do
    [ -f "$tui_file" ] || continue
    . "$tui_file"
done

clashtui() {
    case "${1:-}" in
    -h | --help)
        _tui_help
        ;;
    *)
        _tui_main "$@"
        ;;
    esac
}
