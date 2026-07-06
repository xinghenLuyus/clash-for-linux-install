#!/usr/bin/env bash

api_base() {
    _detect_ext_addr
    printf 'http://127.0.0.1:%s' "$EXT_PORT"
}

api_urlencode() {
    local out='' byte
    for byte in $(printf '%s' "$1" | od -An -tx1 -v); do
        byte=$(printf '%s' "$byte" | tr '[:lower:]' '[:upper:]')
        case $byte in
        2D | 2E | 5F | 7E | 3[0-9] | 4[1-9A-F] | 5[0-9A] | 6[1-9A-F] | 7[0-9A])
            out+="$(printf '%b' "\\x$byte")"
            ;;
        *)
            out+="%$byte"
            ;;
        esac
    done
    printf '%s' "$out"
}

api_curl() {
    local method=$1 path=$2
    shift 2

    local secret base auth=()
    base=$(api_base)
    secret=$(_get_secret)
    [ -n "$secret" ] && auth=(-H "Authorization: Bearer $secret")

    curl -s --noproxy '*' --max-time "${CLASHCTL_API_TIMEOUT:-10}" \
        --location \
        -X "$method" "${auth[@]}" "${base}${path}" "$@"
}

api_get() {
    api_curl GET "$@"
}

api_delete() {
    api_curl DELETE "$@"
}

api_post() {
    api_curl POST "$@"
}

api_put_json() {
    local path=$1 body=$2
    api_curl PUT "$path" -H 'Content-Type: application/json' --data-raw "$body"
}

api_patch_json() {
    local path=$1 body=$2
    api_curl PATCH "$path" -H 'Content-Type: application/json' --data-raw "$body"
}

api_available() {
    service_is_active >&/dev/null || return 1
    local body
    body=$(api_get /version 2>/dev/null) || return 1
    printf '%s' "$body" | "$BIN_YQ" -p json -e '
      has("version") or has("meta") or has("premium")
    ' >/dev/null 2>&1
}

api_configs() {
    api_get /configs
}

api_proxies() {
    api_get /proxies
}

api_proxy_providers() {
    api_get /providers/proxies
}

api_proxy() {
    local enc
    enc=$(api_urlencode "$1")
    api_get "/proxies/$enc"
}

api_select_proxy() {
    local group=$1 proxy=$2 enc body
    enc=$(api_urlencode "$group")
    body=$(NODE=$proxy "$BIN_YQ" -n -o=json '{"name": strenv(NODE)}')
    api_curl PUT "/proxies/$enc" \
        -H 'Content-Type: application/json' \
        --data-raw "$body" \
        -o /dev/null \
        -w '%{http_code}'
}

api_proxy_delay() {
    local proxy=$1 url=$2 timeout=$3 enc qs
    enc=$(api_urlencode "$proxy")
    qs="timeout=${timeout}&url=$(api_urlencode "$url")"
    api_get "/proxies/$enc/delay?$qs"
}

api_group_delay_with_code() {
    local group=$1 url=$2 timeout=$3 enc qs
    enc=$(api_urlencode "$group")
    qs="timeout=${timeout}&url=$(api_urlencode "$url")"
    api_curl GET "/group/$enc/delay?$qs" -w $'\n%{http_code}'
}

api_upgrade() {
    local channel=${1:-}
    api_post "/upgrade?channel=$channel"
}

api_proxy_groups_tsv() {
    api_proxies | "$BIN_YQ" -p json '
      .proxies as $p |
      $p | to_entries[] |
      select(.key != "GLOBAL" and (.value.all != null)) |
      [.key, (.value.type // ""), (.value.now // "")] | @tsv
    ' 2>/dev/null
}

api_proxy_leaf_tsv() {
    api_proxies | "$BIN_YQ" -p json '
      .proxies as $p |
      (
        [($p.DIRECT // empty), ($p.REJECT // empty)] +
        [
          $p | to_entries[] |
          select(.value.all == null and .key != "DIRECT" and .key != "REJECT") |
          .value
        ]
      )[] |
      select(. != null) |
      [.name, (.type // ""), (.provider // ""), (.udp // false), (.xudp // false), (.tfo // false), (.mptcp // false), (.smux // false)] | @tsv
    ' 2>/dev/null
}

api_proxy_provider_leaf_tsv() {
    api_proxy_providers | "$BIN_YQ" -p json '
      (.providers // {}) | to_entries | sort_by(.key) | .[] |
      select(.value.vehicleType == "HTTP" or .value.vehicleType == "File") as $provider |
      ($provider.value.proxies // [])[] |
      [.name, (.type // ""), $provider.key, (.udp // false), (.xudp // false), (.tfo // false), (.mptcp // false), (.smux // false)] | @tsv
    ' 2>/dev/null
}
