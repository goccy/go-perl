package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// surveyNetwork is the Docker network the survey containers run on, the
// default bridge: netguard.sh keeps it from the LAN, the VM's host and the
// VM, and a container is taken off it before its test phase.
const surveyNetwork = "bridge"

// netguardImage is the network helper's image, built from the embedded
// Dockerfile.netguard.
const netguardImage = "cpancompat-netguard"

// guardNetwork puts the rules in place that keep the survey containers from
// the LAN, the VM's host and the VM itself, letting them reach the internet
// and, by DNS only, the resolver they are given (netguard.sh), and verifies
// them. It runs the helper container, with the host's network and
// CAP_NET_ADMIN; neither it nor its image runs CPAN code.
func guardNetwork(dataDir string) error {
	ctx := filepath.Join(dataDir, "netguard-context")
	if err := writeContext(ctx, "Dockerfile.netguard", "netguard.sh"); err != nil {
		return err
	}
	if out, err := exec.Command("docker", "build", "-q", "-t", netguardImage,
		"-f", filepath.Join(ctx, "Dockerfile.netguard"), ctx).CombinedOutput(); err != nil {
		return fmt.Errorf("network helper image: %v: %s", err, out)
	}
	raw, err := exec.Command("docker", "network", "inspect", surveyNetwork).Output()
	if err != nil {
		return fmt.Errorf("docker network inspect %s: %w", surveyNetwork, err)
	}
	var networks []struct {
		Options map[string]string `json:"Options"`
	}
	if err := json.Unmarshal(raw, &networks); err != nil {
		return fmt.Errorf("docker network inspect %s: %w", surveyNetwork, err)
	}
	if len(networks) != 1 || networks[0].Options["com.docker.network.bridge.name"] == "" {
		return fmt.Errorf("docker network inspect %s: no bridge interface", surveyNetwork)
	}
	bridge := networks[0].Options["com.docker.network.bridge.name"]
	// The resolver as a container on the network is given it.
	conf, err := exec.Command("docker", "run", "--rm", "--network", surveyNetwork,
		"--cap-drop", "ALL", "--security-opt", "no-new-privileges",
		netguardImage, "cat", "/etc/resolv.conf").Output()
	if err != nil {
		return fmt.Errorf("the survey network's resolver: %w", err)
	}
	resolvers := nameservers(string(conf))
	if len(resolvers) == 0 {
		return fmt.Errorf("the survey network's resolv.conf names no IPv4 nameserver:\n%s", conf)
	}
	args := append([]string{"run", "--rm", "--network", "host",
		"--cap-drop", "ALL", "--cap-add", "NET_ADMIN", "--security-opt", "no-new-privileges",
		netguardImage, "sh", "/opt/netguard.sh", bridge}, resolvers...)
	out, err := exec.Command("docker", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("network rules: %v: %s", err, out)
	}
	fmt.Fprintf(os.Stderr, "cpancompat: survey containers on %s (%s) reach the internet and DNS at %s only:\n%s",
		surveyNetwork, bridge, strings.Join(resolvers, ", "), out)
	return nil
}

// nameservers returns the IPv4 nameservers a resolv.conf names, in order.
// resolv.conf is resolver(5)'s text format: a "nameserver" line gives one
// address; every other line (comments, options, IPv6 servers, which the
// containers cannot reach) names none.
func nameservers(conf string) []string {
	var out []string
	for _, line := range strings.Split(conf, "\n") {
		f := strings.Fields(line)
		if len(f) < 2 || f[0] != "nameserver" {
			continue
		}
		if ip := net.ParseIP(f[1]); ip != nil && ip.To4() != nil {
			out = append(out, ip.String())
		}
	}
	return out
}

// request is a line dist.sh writes to its standard output to ask the
// driver for something while the container runs.
type request struct {
	Request string `json:"request"`
}

// requestOffline asks that the container be taken off its network before
// its test phase: the tests run with loopback only.
const requestOffline = "offline"

// converse answers the requests dist.sh writes to the container's standard
// output, one JSON object per line, until the container ends; offline is
// how the container is taken off its network. Any other line goes to log as
// it is. dist.sh checks for itself that the request was carried out.
func converse(requests io.Reader, log io.Writer, offline func()) {
	sc := bufio.NewScanner(requests)
	sc.Buffer(make([]byte, 64<<10), 1<<20)
	for sc.Scan() {
		var req request
		if json.Unmarshal(sc.Bytes(), &req) == nil && req.Request == requestOffline {
			offline()
			continue
		}
		log.Write(append(sc.Bytes(), '\n'))
	}
	// A line too long to scan: the rest is output, which must still be
	// read for the container to go on.
	io.Copy(log, requests)
}

// takeOffline disconnects the container name from the survey network and
// then tells dist.sh so (the file it waits for, which only root can make).
// Failures are written to log; dist.sh finds the network still there.
func takeOffline(name string, log io.Writer) {
	if out, err := exec.Command("docker", "network", "disconnect", "--force", surveyNetwork, name).CombinedOutput(); err != nil {
		fmt.Fprintf(log, "cpancompat: docker network disconnect: %v: %s\n", err, out)
	}
	if out, err := exec.Command("docker", "exec", "--user", "root", name, "touch", "/run/cpancompat/offline").CombinedOutput(); err != nil {
		fmt.Fprintf(log, "cpancompat: docker exec: %v: %s\n", err, out)
	}
}
