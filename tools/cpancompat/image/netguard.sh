#!/bin/sh
# netguard.sh <bridge interface> <resolver>...
#
# Runs in the network helper container (Dockerfile.netguard: no CPAN code,
# the host's network, CAP_NET_ADMIN) when the driver starts a run: keeps the
# survey containers, which run code from CPAN that may be malicious, from
# reaching the machine's LAN, the host the VM runs on, and the VM itself.
# They reach the internet (Alien distributions download what they build) and
# the resolver they are given, by DNS only:
#
# - Forwarded traffic from the bridge (the DOCKER-USER chain, which Docker
#   leaves to its user) to private (10/8, 172.16/12, 192.168/16),
#   link-local (169.254/16) and shared (100.64/10) addresses is rejected:
#   that covers the LAN, the VM's host (192.168.5.2 here) and the other
#   containers.
# - Traffic from the bridge to the VM itself (the INPUT chain: the bridge's
#   gateway, the VM's own addresses) is rejected.
#
# The rules are in chains of their own, CPANCOMPAT-FORWARD and
# CPANCOMPAT-INPUT, jumped to from DOCKER-USER and INPUT for traffic coming
# in on the bridge. When they are already in place nothing changes;
# otherwise new chains are filled and swapped in, so that containers already
# running are never left without them. It exits 0 only when every rule is
# verified in place.
#
# iptables is the nf_tables one, as Docker's.
set -u
bridge=$1
shift
resolvers="$*"
blocked="10.0.0.0/8 172.16.0.0/12 192.168.0.0/16 169.254.0.0/16 100.64.0.0/10"

# rules <chain kind>: the rules of a chain, one per line, as iptables
# arguments.
rules() {
    for r in $resolvers; do
        echo "-d $r/32 -p udp -m udp --dport 53 -j RETURN"
        echo "-d $r/32 -p tcp -m tcp --dport 53 -j RETURN"
    done
    case $1 in
    FORWARD)
        for net in $blocked; do
            echo "-d $net -j REJECT --reject-with icmp-port-unreachable"
        done
        ;;
    INPUT)
        echo "-j REJECT --reject-with icmp-port-unreachable"
        ;;
    esac
}

# parent <chain kind>: the built-in or Docker chain that jumps to ours.
parent() {
    case $1 in
    FORWARD) echo DOCKER-USER ;;
    INPUT) echo INPUT ;;
    esac
}

# in_place <chain kind>: whether CPANCOMPAT-<kind> holds exactly its rules
# and is jumped to for the bridge's traffic.
in_place() {
    chain=CPANCOMPAT-$1
    iptables -C "$(parent "$1")" -i "$bridge" -j "$chain" 2> /dev/null || return 1
    n=0
    while read -r rule; do
        # shellcheck disable=SC2086  # a rule is a list of arguments
        iptables -C "$chain" $rule 2> /dev/null || return 1
        n=$((n + 1))
    done <<EOF
$(rules "$1")
EOF
    # iptables -S lists the chain itself, then one line per rule.
    test "$(iptables -S "$chain" | wc -l)" -eq $((n + 1))
}

# install <chain kind>: fills a new chain and swaps it in for the old one.
install() {
    chain=CPANCOMPAT-$1
    new=CPANCOMPAT-NEW-$$
    iptables -N "$new" || return 1
    while read -r rule; do
        # shellcheck disable=SC2086
        iptables -A "$new" $rule || return 1
    done <<EOF
$(rules "$1")
EOF
    iptables -I "$(parent "$1")" 1 -i "$bridge" -j "$new" || return 1
    while iptables -D "$(parent "$1")" -i "$bridge" -j "$chain" 2> /dev/null; do :; done
    if iptables -S "$chain" > /dev/null 2>&1; then
        iptables -F "$chain" && iptables -X "$chain" || return 1
    fi
    iptables -E "$new" "$chain"
}

test -n "$resolvers" || { echo "netguard.sh: no resolver" >&2; exit 2; }
iptables -S DOCKER-USER > /dev/null || { echo "netguard.sh: no DOCKER-USER chain in this iptables" >&2; exit 2; }
for kind in FORWARD INPUT; do
    if ! in_place $kind; then
        echo "netguard.sh: installing CPANCOMPAT-$kind"
        install $kind || { echo "netguard.sh: installing CPANCOMPAT-$kind failed" >&2; exit 1; }
    fi
    in_place $kind || { echo "netguard.sh: CPANCOMPAT-$kind not in place" >&2; exit 1; }
    iptables -S "CPANCOMPAT-$kind"
done
