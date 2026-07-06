#!/usr/bin/env bash

_overview_runtime_value() {
    "$BIN_YQ" -r "$1 // \"\"" "$CLASH_CONFIG_RUNTIME" 2>/dev/null
}

_overview_live_value() {
    local live_json=$1 expr=$2 value
    [ -n "$live_json" ] || return 1
    value=$("$BIN_YQ" -p json -r "$expr // \"\"" <<<"$live_json" 2>/dev/null)
    [ "$value" = null ] && value=
    [ -n "$value" ] || return 1
    printf '%s' "$value"
}

_overview_value() {
    local live_json=$1 live_expr=$2 runtime_expr=$3 fallback=${4:-未设置}
    local value
    value=$(_overview_live_value "$live_json" "$live_expr") && {
        printf '%s' "$value"
        return
    }
    value=$(_overview_runtime_value "$runtime_expr")
    printf '%s' "${value:-$fallback}"
}

_overview_secret() {
    [ -x "$BIN_YQ" ] || {
        printf ''
        return 0
    }
    _get_secret 2>/dev/null
}

_tui_overview() {
    _tui_status_block
}

_tui_status_block() {
    local service_state api_state tun_state profile secret live_json
    local mode mixed_port http_port socks_port ext allow_lan bind_addr dns_enabled

    service_state=stopped
    service_is_active >&/dev/null && service_state=running

    api_state=unavailable
    if [ "$service_state" = running ]; then
        live_json=$(api_configs 2>/dev/null)
        [ -n "$live_json" ] && api_state=available
    fi

    tun_state=off
    tunstatus >&/dev/null && tun_state=on

    mode=$(_overview_value "$live_json" '.mode' '.mode' rule)
    mixed_port=$(_overview_value "$live_json" '."mixed-port"' '."mixed-port"')
    http_port=$(_overview_value "$live_json" '.port' '.port')
    socks_port=$(_overview_value "$live_json" '."socks-port"' '."socks-port"')
    ext=$(_overview_value "$live_json" '."external-controller"' '."external-controller"')
    allow_lan=$(_overview_value "$live_json" '."allow-lan"' '."allow-lan"' false)
    bind_addr=$(_overview_value "$live_json" '."bind-address"' '."bind-address"' 127.0.0.1)
    dns_enabled=$(_overview_value "$live_json" '.dns.enable' '.dns.enable')

    secret=$(_overview_secret)
    profile=$(_tui_current_profile)

    printf 'clashctl TUI\n'
    printf '\n'
    printf 'Core\n'
    printf '  Kernel:      %s\n' "$CLASHCTL_KERNEL"
    printf '  Service:     %s\n' "$service_state"
    printf '  API:         %s\n' "$api_state"
    printf '  Mode:        %s\n' "$mode"
    printf '  Controller:  %s\n' "$ext"
    printf '\n'
    printf 'Proxy\n'
    printf '  Mixed:       %s\n' "$mixed_port"
    printf '  HTTP:        %s\n' "$http_port"
    printf '  SOCKS:       %s\n' "$socks_port"
    printf '  Allow LAN:   %s\n' "$allow_lan"
    printf '  Bind:        %s\n' "$bind_addr"
    printf '\n'
    printf 'Runtime\n'
    printf '  TUN:         %s\n' "$tun_state"
    printf '  DNS:         %s\n' "$dns_enabled"
    printf '  Secret:      %s\n' "$(_tui_mask_secret "$secret")"
    printf '  Profile:     %s\n' "$profile"

    [ "$api_state" = available ] && drift_print "$live_json"
    return 0
}
