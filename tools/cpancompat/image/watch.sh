#!/bin/bash
# watch.sh <parent pid> <limit seconds> <log> <max log bytes> <out prefix>
#
# Watches one phase of dist.sh (the /usr/bin/timeout process that is a
# child of <parent pid>) until dist.sh stops it with SIGTERM:
#
# - Every five seconds, the size of the phase's log. Past <max log bytes>
#   (a script that prompts in a loop on a closed stdin, a test that prints
#   without end) the phase is ended, and <out prefix>.output-limit says so.
# - Fifteen seconds before the limit, what every process of the container
#   is doing goes to <out prefix>.stuck.jsonl, one JSON object per process:
#   its pid, parent, state, how long it has been running, the CPU it used
#   in a five-second window, the kernel function it and each of its threads
#   wait in, its arguments and its descriptors. A process that waits and
#   used no CPU is blocked; one that used CPU while its test made no
#   progress is spinning.
#
# Written by the shell on purpose, like dist.sh: it must not depend on the
# perl under measurement.
#
# SIGTERM ends it by its default action. A trap would not do: bash runs a
# trap where the signal finds it, and when that is in the middle of parsing
# a command substitution, bash 5.2 fails to parse the trap's own text
# ("unexpected EOF while looking for matching `)'") and goes on watching, so
# dist.sh would wait for it until the container is ended. The record of the
# processes is written under another name and renamed when it is whole,
# and dist.sh lets it finish before stopping the watcher.
parent=$1 limit=$2 log=$3 max=$4 prefix=$5

# The phase's /usr/bin/timeout process.
phase_pid() {
    local d st
    for d in /proc/[0-9]*; do
        read -r -a st < "$d/stat" 2> /dev/null || continue
        if [ "${st[3]}" = "$parent" ] && [ "${st[1]}" = "(timeout)" ]; then
            echo "${d#/proc/}"
            return
        fi
    done
}

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

processes() {
    local hz uptime row p pp s t start args line sep a f k w
    hz=$(getconf CLK_TCK)
    read -r uptime _ < /proc/uptime
    declare -A before
    while read -r p _ _ t _; do before[$p]=$t; done < <(snap)
    sleep 5
    # Taken whole before the loop below starts the processes of its own.
    local rows
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
    done
}

start=$SECONDS
snapshot_at=$(( limit > 30 ? limit - 15 : limit ))
taken=false
while :; do
    sleep 5
    size=$(stat -c %s "$log" 2> /dev/null || echo 0)
    if [ "$size" -gt "$max" ]; then
        printf '{"bytes":%d,"limit":%d}\n' "$size" "$max" > "$prefix.output-limit"
        pid=$(phase_pid)
        [ -n "$pid" ] && kill -TERM "$pid"
        exit 0
    fi
    if ! $taken && [ $((SECONDS - start)) -ge "$snapshot_at" ]; then
        processes > "$prefix.stuck.jsonl.partial"
        mv "$prefix.stuck.jsonl.partial" "$prefix.stuck.jsonl"
        taken=true
    fi
done
