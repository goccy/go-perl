#!/bin/bash
# stuck.sh <seconds> <file>: after <seconds>, writes to <file> what every
# process of the container is doing, for a phase that is about to run out of
# time. A phase that ends first stops this script with SIGTERM.
#
# One JSON object per process: its pid, parent, state, how long it has been
# running, the CPU it used in a five-second window, the kernel function it
# and each of its threads wait in, its arguments and its descriptors. A
# process that waits and used no CPU is blocked; one that used CPU while its
# test made no progress is spinning.
#
# Written by the shell on purpose, like dist.sh: it must not depend on the
# perl under measurement.
trap 'kill "$sleeper" 2> /dev/null; exit 0' TERM
sleep "$1" &
sleeper=$!
wait "$sleeper"

# A JSON string of $1. Control characters other than tab, newline and
# carriage return are dropped: they cannot appear in a JSON string as such.
str() {
    local s
    s=$(printf '%s' "$1" | tr -d '\000-\010\013\014\016-\037')
    s=${s//\\/\\\\}
    s=${s//\"/\\\"}
    s=${s//$'\t'/\\t}
    s=${s//$'\n'/\\n}
    s=${s//$'\r'/\\r}
    printf '"%s"' "$s"
}

snap() {
    local d p st
    for d in /proc/[0-9]*; do
        p=${d#/proc/}
        read -r -a st < "$d/stat" 2> /dev/null || continue
        echo "$p ${st[3]} ${st[2]} $((st[13] + st[14])) ${st[21]}"
    done
}

hz=$(getconf CLK_TCK)
read -r uptime _ < /proc/uptime
declare -A before
while read -r p _ _ t _; do before[$p]=$t; done < <(snap)
sleep 5
# Taken whole before the loop below starts the processes of its own.
mapfile -t rows < <(snap)
for row in "${rows[@]}"; do
    read -r p pp s t start <<< "$row"
    [ "$p" = 1 ] && continue
    [ "$p" = "$$" ] && continue
    [ "$pp" = "$$" ] && continue
    args=()
    mapfile -d '' args < "/proc/$p/cmdline" 2> /dev/null || continue
    [ ${#args[@]} -gt 0 ] || continue
    line="{\"pid\":$p,\"ppid\":$pp,\"state\":$(str "$s")"
    line+=",\"elapsed_seconds\":$(awk -v u="$uptime" -v s="$start" -v hz="$hz" 'BEGIN { printf "%d", u - s / hz + 5 }')"
    line+=",\"cpu_percent\":$(( (t - ${before[$p]:-$t}) * 100 / (hz * 5) ))"
    line+=",\"wchan\":$(str "$(cat "/proc/$p/wchan" 2> /dev/null)")"
    sep=
    line+=',"args":['
    for a in "${args[@]}"; do line+="$sep$(str "$a")"; sep=,; done
    line+='],"fds":{'
    sep=
    for f in "/proc/$p/fd/"*; do
        [ -e "$f" ] || [ -L "$f" ] || continue
        line+="$sep$(str "${f##*/}"):$(str "$(readlink "$f")")"
        sep=,
    done
    line+='},"thread_wchans":{'
    sep=
    declare -A waits=()
    for w in /proc/$p/task/*/wchan; do
        k=$(cat "$w" 2> /dev/null) || continue
        waits[$k]=$(( ${waits[$k]:-0} + 1 ))
    done
    for k in "${!waits[@]}"; do line+="$sep$(str "$k"):${waits[$k]}"; sep=,; done
    unset waits
    line+='}}'
    echo "$line"
done > "$2"
