#!/usr/bin/env bash

proxy_default_delay_url() {
    printf '%s' "${CLASHCTL_NODE_DELAY_URL:-http://www.gstatic.com/generate_204}"
}

proxy_default_delay_timeout() {
    printf '%s' "${CLASHCTL_NODE_DELAY_TIMEOUT:-5000}"
}

proxy_validate_delay_url() {
    [[ $1 =~ ^https?:// ]] || {
        _errorcat "测速 URL 需以 http:// 或 https:// 开头：$1"
        return 1
    }
}

proxy_validate_delay_timeout() {
    [[ $1 =~ ^[1-9][0-9]*$ ]] || {
        _errorcat "测速超时时间必须为正整数毫秒：$1"
        return 1
    }
}

proxy_groups_tsv() {
    api_proxies | "$BIN_YQ" -p json '
      .proxies as $p |
      ($p.GLOBAL.all // []) as $globalNames |
      ([ $p | to_entries[] | select(.key != "GLOBAL" and (.value.all != null)) | .key ]) as $allGroups |
      ([ $globalNames[] | select($p[.] != null and $p[.].all != null) ]) as $globalGroups |
      (
        ($allGroups | map(select(. as $name | ($globalGroups | index($name) | not)))) +
        $globalGroups
      )[] |
      [., ($p[.].type // ""), ($p[.].now // ""), (($p[.].all // []) | length)] | @tsv
    ' 2>/dev/null
}

proxy_leaf_tsv() {
    {
        api_proxy_leaf_tsv
        api_proxy_provider_leaf_tsv
    } | awk -F '\t' '
      !seen[$1]++ {
        caps=""
        if ($4 == "true") caps=caps " UDP"
        if ($5 == "true") caps=caps " XUDP"
        if ($6 == "true") caps=caps " TFO"
        if ($7 == "true") caps=caps " MPTCP"
        if ($8 == "true") caps=caps " SMUX"
        sub(/^ /, "", caps)
        print $1 "\t" $2 "\t" $3 "\t" caps
      }'
}

proxy_group_json() {
    api_proxy "$1"
}

proxy_group_members() {
    proxy_group_json "$1" | "$BIN_YQ" -p json '.all // [] | .[]' 2>/dev/null
}

proxy_group_current() {
    proxy_group_json "$1" | "$BIN_YQ" -p json '.now // ""' 2>/dev/null
}

proxy_classify() {
    local exists is_group
    IFS=$'\t' read -r exists is_group < <(
        api_proxies | PROXY_NAME=$1 "$BIN_YQ" -p json '
            [(.proxies[strenv(PROXY_NAME)] != null),
             (.proxies[strenv(PROXY_NAME)].all != null)] | @tsv' 2>/dev/null
    )
    if [ "$exists" != true ]; then
        printf 'none'
    elif [ "$is_group" = true ]; then
        printf 'group'
    else
        printf 'proxy'
    fi
}

proxy_member_rows() {
    local group=$1 filter=${2:-} sort=${3:-default}
    local now name type provider caps marker
    now=$(proxy_group_current "$group")

    declare -A type_map provider_map caps_map delay_map
    while IFS=$'\t' read -r name type provider caps; do
        [ -z "$name" ] && continue
        type_map["$name"]=$type
        provider_map["$name"]=$provider
        caps_map["$name"]=$caps
    done < <(proxy_leaf_tsv)

    while IFS= read -r name; do
        [ -z "$name" ] && continue
        [ -n "$filter" ] && [[ "$name" != *"$filter"* ]] && continue
        marker=' '
        [ "$name" = "$now" ] && marker='*'
        printf '%s\t%s\t%s\t%s\t%s\t%s\n' \
            "$marker" "$name" "${type_map[$name]:-unknown}" "${provider_map[$name]:-}" "${caps_map[$name]:-}" "${delay_map[$name]:-}"
    done < <(proxy_group_members "$group") | proxy_sort_member_rows "$sort"
}

proxy_sort_member_rows() {
    case "${1:-default}" in
    name)
        sort -t $'\t' -k2,2
        ;;
    delay)
        sort -t $'\t' -k6,6n -k2,2
        ;;
    default | *)
        cat
        ;;
    esac
}

proxy_select() {
    api_select_proxy "$@"
}

proxy_apply() {
    local group=$1 member=$2 code
    code=$(proxy_select "$group" "$member")
    case $code in
    204) _okcat "已切换 [$group] → $member" ;;
    400) _failcat "切换失败：节点 [$member] 不在策略组 [$group] 内，或该组不可手动切换" ;;
    404) _failcat "切换失败：策略组 [$group] 不存在" ;;
    *) _failcat "切换失败：控制器返回 $code（检查内核是否运行 / secret 是否正确）" ;;
    esac
}

proxy_delay_one_row() {
    local name=$1 url=$2 timeout=$3 resp delay
    resp=$(api_proxy_delay "$name" "$url" "$timeout")
    delay=$("$BIN_YQ" -p json '.delay // ""' <<<"$resp" 2>/dev/null)
    printf '%s\t%s\n' "$name" "$delay"
}

proxy_delay_group_rows() {
    local group=$1 url=$2 timeout=$3 resp code body name delay
    resp=$(api_group_delay_with_code "$group" "$url" "$timeout")
    code=${resp##*$'\n'}
    body=${resp%$'\n'*}

    if [ "$code" = 200 ]; then
        "$BIN_YQ" -p json 'to_entries | .[] | [.key, .value] | @tsv' <<<"$body" 2>/dev/null
        return 0
    fi

    local concurrency active=0
    concurrency=${CLASHCTL_NODE_DELAY_CONCURRENCY:-8}
    [[ "$concurrency" =~ ^[0-9]+$ ]] || concurrency=8
    ((concurrency < 1)) && concurrency=1

    {
        while IFS= read -r name; do
            [ -z "$name" ] && continue
            proxy_delay_one_row "$name" "$url" "$timeout" &
            ((active += 1))
            if ((active >= concurrency)); then
                wait
                active=0
            fi
        done < <(proxy_group_members "$group")
        wait
    }
}

proxy_print_delay_rows() {
    local name delay
    while IFS=$'\t' read -r name delay; do
        if [[ "$delay" =~ ^[0-9]+$ ]] && [ "$delay" -gt 0 ]; then
            printf '%010d\t%s\t%sms\n' "$delay" "$name" "$delay"
        else
            printf '9999999999\t%s\ttimeout\n' "$name"
        fi
    done | sort -n | while IFS=$'\t' read -r _ name delay; do
        printf '  %-36s %s\n' "$name" "$delay"
    done
}

proxy_print_groups() {
    local name type now count
    while IFS=$'\t' read -r name type now count; do
        printf '  %-24s → %-28s [%s, %s]\n' "$name" "${now:-—}" "${type:-unknown}" "${count:-0}"
    done < <(proxy_groups_tsv)
}

proxy_print_members() {
    local group=$1 filter=${2:-} sort=${3:-default}
    local marker name type provider caps delay meta
    while IFS=$'\t' read -r marker name type provider caps delay; do
        meta=$type
        [ -n "$provider" ] && meta="$meta@$provider"
        [ -n "$caps" ] && meta="$meta $caps"
        printf '  %s %-36s [%s]\n' "$marker" "$name" "$meta"
    done < <(proxy_member_rows "$group" "$filter" "$sort")
}

proxy_pick_group() {
    local selected
    if command -v fzf >&/dev/null && [ -t 0 ]; then
        selected=$(proxy_groups_tsv | fzf \
            --height=100% \
            --layout=reverse \
            --border \
            --delimiter=$'\t' \
            --with-nth='1,3,2,4' \
            --prompt='proxies > ' \
            --header='选择策略组 | Enter 确认 | q/Esc 退出' \
            --preview='bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && proxy_preview_group \"$1\"" -- {1}' \
            --preview-window='right:60%:wrap' \
            --bind='q:abort,esc:abort,ctrl-c:abort') || return
        printf '%s\n' "${selected%%$'\t'*}"
        return
    fi

    proxy_print_groups >&2
    printf '%s' "$(_okcat '✈️ ' '请输入策略组名称：')" >&2
    read -r selected
    [ -n "$selected" ] && printf '%s\n' "$selected"
}

proxy_pick_member() {
    local group=$1 selected
    if command -v fzf >&/dev/null && [ -t 0 ]; then
        selected=$(proxy_member_rows "$group" | fzf \
            --height=100% \
            --layout=reverse \
            --border \
            --delimiter=$'\t' \
            --with-nth='1,2,3,4,5' \
            --prompt="$group > " \
            --header='选择节点 | Enter 切换 | q/Esc 返回' \
            --preview='bash -c ". \"$CLASHCTL_HOME/scripts/cmd/clashctl.sh\" && proxy_preview_member \"$1\" \"$2\"" -- "$group" {2}' \
            --preview-window='right:50%:wrap' \
            --bind='q:abort,esc:abort,ctrl-c:abort') || return
        printf '%s\n' "$(printf '%s' "$selected" | awk -F '\t' '{print $2}')"
        return
    fi

    proxy_print_members "$group" >&2
    printf '%s' "$(_okcat '✈️ ' '请输入节点名称：')" >&2
    read -r selected
    [ -n "$selected" ] && printf '%s\n' "$selected"
}

proxy_preview_group() {
    local group=$1 type now count
    IFS=$'\t' read -r _ type now count < <(proxy_groups_tsv | awk -F '\t' -v g="$group" '$1 == g { print; exit }')
    printf '策略组\n'
    printf '  名称：%s\n' "$group"
    printf '  类型：%s\n' "${type:-unknown}"
    printf '  当前：%s\n' "${now:-—}"
    printf '  节点：%s\n\n' "${count:-0}"
    printf '节点\n'
    proxy_print_members "$group" | sed -n '1,80p'
}

proxy_preview_member() {
    local group=$1 member=$2 marker name type provider caps delay meta
    while IFS=$'\t' read -r marker name type provider caps delay; do
        [ "$name" = "$member" ] || continue
        meta=$type
        [ -n "$provider" ] && meta="$meta@$provider"
        [ -n "$caps" ] && meta="$meta $caps"
        printf '节点\n'
        printf '  名称：%s\n' "$name"
        printf '  策略组：%s\n' "$group"
        printf '  类型：%s\n' "$type"
        printf '  Provider：%s\n' "${provider:-—}"
        printf '  能力：%s\n' "${caps:-—}"
        [ "$marker" = '*' ] && printf '  状态：当前选中\n' || printf '  状态：可切换\n'
        return
    done < <(proxy_member_rows "$group")
}

proxy_cli_node() {
    case "${1:-}" in
    -h | --help | help)
        proxy_node_help
        return 0
        ;;
    esac

    service_is_active >&/dev/null || {
        _failcat "$CLASHCTL_KERNEL 未运行，请先执行 clashctl on"
        return 1
    }

    case "${1:-}" in
    ls | list)
        shift
        proxy_cli_list "$@"
        ;;
    use)
        shift
        proxy_cli_use "$@"
        ;;
    delay)
        shift
        proxy_cli_delay "$@"
        ;;
    -* | '')
        proxy_cli_use "$@"
        ;;
    *)
        _errorcat "未知 node 子命令：$1"
        proxy_node_help
        return 1
        ;;
    esac
}

proxy_cli_list() {
    local filter= sort=default args=()
    while [ $# -gt 0 ]; do
        case "$1" in
        --filter)
            filter=$2
            shift
            ;;
        --filter=*)
            filter=${1#*=}
            ;;
        --sort)
            sort=$2
            shift
            ;;
        --sort=*)
            sort=${1#*=}
            ;;
        *)
            args+=("$1")
            ;;
        esac
        shift
    done

    if [ ${#args[@]} -eq 0 ]; then
        proxy_print_groups
    else
        proxy_print_members "${args[0]}" "$filter" "$sort"
    fi
}

proxy_cli_use() {
    local group=${1:-} member=${2:-}
    [ -z "$group" ] && group=$(proxy_pick_group) || true
    [ -z "$group" ] && return 1
    [ -z "$member" ] && member=$(proxy_pick_member "$group") || true
    [ -z "$member" ] && return 1
    proxy_apply "$group" "$member"
}

proxy_cli_delay() {
    local mode=auto url timeout target args=()
    url=$(proxy_default_delay_url)
    timeout=$(proxy_default_delay_timeout)

    while [ $# -gt 0 ]; do
        case "$1" in
        -g | --group) mode=group ;;
        -p | --proxy) mode=proxy ;;
        -u | --url)
            url=$2
            shift
            ;;
        --url=*) url=${1#*=} ;;
        -t | --timeout)
            timeout=$2
            shift
            ;;
        --timeout=*) timeout=${1#*=} ;;
        *) args+=("$1") ;;
        esac
        shift
    done
    proxy_validate_delay_url "$url" || return
    proxy_validate_delay_timeout "$timeout" || return

    target=${args[0]:-}
    if [ -z "$target" ]; then
        target=$(proxy_pick_group) || return
        mode=group
    fi

    case "$mode:$target" in
    proxy:*)
        proxy_delay_one_row "$target" "$url" "$timeout" | proxy_print_delay_rows
        ;;
    group:* | auto:*)
        if [ "$mode" = auto ] && [ "$(proxy_classify "$target")" = proxy ]; then
            proxy_delay_one_row "$target" "$url" "$timeout" | proxy_print_delay_rows
        else
            proxy_delay_group_rows "$target" "$url" "$timeout" | proxy_print_delay_rows
        fi
        ;;
    esac
}

proxy_node_help() {
    cat <<EOF

clashctl node - Clash Verge 风格节点管理

Usage:
  clashctl node                 交互式选择策略组与节点
  clashctl node ls [策略组]      列出策略组或组内节点
  clashctl node use [组] [节点]  切换节点
  clashctl node delay [名称]     测速策略组或节点

Options:
  --filter <文本>               过滤节点，配合 ls <策略组>
  --sort default|delay|name      排序节点，配合 ls <策略组>
  -g, --group                   delay 强制按策略组测速
  -p, --proxy                   delay 强制按节点测速
  -u, --url <URL>               测速 URL
  -t, --timeout <毫秒>          测速超时

EOF
}
