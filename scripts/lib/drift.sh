#!/usr/bin/env bash

drift_fields() {
    cat <<EOF
mode|.mode
mixed-port|."mixed-port"
port|.port
socks-port|."socks-port"
allow-lan|."allow-lan"
bind-address|."bind-address"
ipv6|.ipv6
log-level|."log-level"
external-controller|."external-controller"
tun.enable|.tun.enable
dns.enable|.dns.enable
EOF
}

drift_value() {
    local format=$1 expr=$2 source=$3
    local value
    if [ "$format" = json ]; then
        value=$("$BIN_YQ" -p json -r "$expr // \"\"" <<<"$source" 2>/dev/null)
    else
        value=$("$BIN_YQ" -r "$expr // \"\"" "$source" 2>/dev/null)
    fi
    case "$value" in
    null) value= ;;
    esac
    printf '%s' "$value"
}

drift_rows() {
    local live_json=$1
    [ -n "$live_json" ] || return 0

    local key expr runtime_value live_value
    while IFS='|' read -r key expr; do
        [ -n "$key" ] || continue
        runtime_value=$(drift_value yaml "$expr" "$CLASH_CONFIG_RUNTIME")
        live_value=$(drift_value json "$expr" "$live_json")
        [ -z "$live_value" ] && continue
        [ "$runtime_value" = "$live_value" ] && continue
        printf '%s\t%s\t%s\n' "$key" "${runtime_value:-unset}" "$live_value"
    done < <(drift_fields)
}

drift_print() {
    local live_json=$1 rows
    rows=$(drift_rows "$live_json")
    [ -n "$rows" ] || return 0

    printf '\n配置提示\n'
    printf '检测到运行态配置与 runtime.yaml 不一致：\n\n'
    printf '%s\n' "$rows" | while IFS=$'\t' read -r key runtime_value live_value; do
        printf '  %s\n' "$key"
        printf '    runtime.yaml: %s\n' "$runtime_value"
        printf '    live core:    %s\n' "$live_value"
    done
    printf '\n说明：运行态配置可能由 Web UI 或外部 API 临时修改；重启服务、切换订阅、编辑 Mixin 或重新合并配置后，将以 mixin.yaml/runtime.yaml 为准。\n'
}
