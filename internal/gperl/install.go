package gperl

// The toolchain gperl manages.
//
// gperl build needs the Go toolchain, and building an XS distribution
// needs a C compiler. `gperl install toolchain` puts a pinned release of
// each — Go, and Zig as the C compiler (it also cross compiles) — into the
// toolchain directory (GPERLTOOLCHAIN, default ~/gperl/toolchain), each
// download checked against the hash recorded here.
//
// Which tool a command uses: one the environment names explicitly (CC, AR)
// wins, then the managed one when installed, then the one on PATH.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/goccy/go-perl/internal/mod"
)

// managedTool is one pinned tool release.
type managedTool struct {
	name    string
	version string
	// url returns the archive's address for a platform ("darwin_arm64").
	url func(platform string) string
	// sha256 is the archive's hash by platform.
	sha256 map[string]string
	// bin is the tool's executable below its install directory.
	bin string
}

var managedTools = []managedTool{
	{
		name: "go", version: "1.26.4", bin: filepath.Join("bin", "go"),
		url: func(platform string) string {
			return "https://go.dev/dl/go1.26.4." + strings.Replace(platform, "_", "-", 1) + ".tar.gz"
		},
		sha256: map[string]string{
			"darwin_amd64": "05dc9b5f9997744520aaebb3d5deaa7c755371aebbfb7f97c2511a9f3367538d",
			"darwin_arm64": "b62ad2b6d7d2464f12a5bcad7ff47f19d08325773b5efd21610e445a05a9bf53",
			"linux_amd64":  "1153d3d50e0ac764b447adfe05c2bcf08e889d42a02e0fe0259bd47f6733ad7f",
			"linux_arm64":  "ef758ae7c6cf9267c9c0ef080b8965f453d89ab2d25d9eb22de4405925238768",
		},
	},
	{
		name: "zig", version: "0.17.1", bin: "zig",
		url: func(platform string) string {
			return "https://ziglang.org/download/0.17.1/zig-" + zigTriple(platform) + "-0.17.1.tar.xz"
		},
		sha256: map[string]string{
			"darwin_amd64": "9af07b20319fbd0c92ee46ad23e81d6a782b1aaea2cfb8dcbff35f1d2c7ddbb7",
			"darwin_arm64": "75f1bc7a1e9852f17ec9a0c1895f767c4ec56d46bc8a834920eece1d0359dd8d",
			"linux_amd64":  "3581863e3cbcfd0fce0efc682ce7e9e93da96668a977791d356a26d2a7e9b17a",
			"linux_arm64":  "a1546d5646eb184437918d6635ca6b762a0e889eac09cef720ee7a1ed3fe7795",
		},
	},
}

// zigTriple spells a platform the way Zig names its downloads and targets
// ("linux_amd64" -> "x86_64-linux").
func zigTriple(platform string) string {
	goos, goarch, _ := strings.Cut(platform, "_")
	cpu := map[string]string{"amd64": "x86_64", "arm64": "aarch64"}[goarch]
	if goos == "darwin" {
		goos = "macos"
	}
	return cpu + "-" + goos
}

// toolchainDir returns the directory the managed tools live in:
// GPERLTOOLCHAIN, or ~/gperl/toolchain.
func toolchainDir() (string, error) {
	if dir := os.Getenv("GPERLTOOLCHAIN"); dir != "" {
		return filepath.Abs(dir)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locate the toolchain directory (set GPERLTOOLCHAIN): %w", err)
	}
	return filepath.Join(home, "gperl", "toolchain"), nil
}

// managedBin returns the executable of the managed tool name, or "" when
// it is not installed.
func managedBin(name string) string {
	dir, err := toolchainDir()
	if err != nil {
		return ""
	}
	for _, tool := range managedTools {
		if tool.name != name {
			continue
		}
		bin := filepath.Join(dir, tool.name+"-"+tool.version, tool.bin)
		if _, err := os.Stat(bin); err == nil {
			return bin
		}
	}
	return ""
}

// hostPlatform names the platform this process runs on.
func hostPlatform() string { return runtime.GOOS + "_" + runtime.GOARCH }

// InstallToolchain implements `gperl install toolchain`: it downloads the
// pinned Go and Zig releases for this platform into the toolchain
// directory. A tool already present is left as it is.
func InstallToolchain() error {
	dir, err := toolchainDir()
	if err != nil {
		return err
	}
	platform := hostPlatform()
	for _, tool := range managedTools {
		dst := filepath.Join(dir, tool.name+"-"+tool.version)
		if _, err := os.Stat(filepath.Join(dst, tool.bin)); err == nil {
			fmt.Fprintf(os.Stderr, "gperl: %s %s is already installed in %s\n", tool.name, tool.version, dst)
			continue
		}
		want, ok := tool.sha256[platform]
		if !ok {
			return fmt.Errorf("no %s release is pinned for %s", tool.name, platform)
		}
		if err := installTool(dir, dst, tool.url(platform), want); err != nil {
			return fmt.Errorf("install %s %s: %w", tool.name, tool.version, err)
		}
		fmt.Fprintf(os.Stderr, "gperl: installed %s %s in %s\n", tool.name, tool.version, dst)
	}
	return nil
}

// installTool downloads the archive at url, checks its hash, and unpacks
// it as dst (without the archive's single top-level directory).
func installTool(dir, dst, url, want string) (err error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	archive, err := os.CreateTemp(dir, "download-*"+filepath.Ext(url))
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.Remove(archive.Name())) }()
	fmt.Fprintf(os.Stderr, "gperl: downloading %s\n", url)
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return errors.Join(err, archive.Close())
	}
	req.Header.Set("User-Agent", "gperl")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Join(err, archive.Close())
	}
	sum := sha256.New()
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("GET %s: %s", url, resp.Status)
	} else {
		_, err = io.Copy(io.MultiWriter(archive, sum), resp.Body)
	}
	if err = errors.Join(err, resp.Body.Close(), archive.Close()); err != nil {
		return err
	}
	if got := hex.EncodeToString(sum.Sum(nil)); got != want {
		return fmt.Errorf("%s: sha256 %s does not match the pinned %s", url, got, want)
	}
	tmp, err := os.MkdirTemp(dir, "unpack-*")
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, os.RemoveAll(tmp)) }()
	if strings.HasSuffix(url, ".tar.gz") {
		err = mod.ExtractTar(archive.Name(), tmp)
	} else {
		// The standard library has no xz decoder; tar on every supported
		// platform reads .tar.xz.
		err = run("tar", "-xf", archive.Name(), "-C", tmp)
	}
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(tmp)
	if err != nil {
		return err
	}
	if len(entries) != 1 || !entries[0].IsDir() {
		return fmt.Errorf("%s: want one top-level directory in the archive, found %d entries", url, len(entries))
	}
	return os.Rename(filepath.Join(tmp, entries[0].Name()), dst)
}

// goCommand returns the go command to run: the managed one when
// installed, the one on PATH otherwise.
func goCommand() (string, error) {
	if bin := managedBin("go"); bin != "" {
		return bin, nil
	}
	if bin, err := exec.LookPath("go"); err == nil {
		return bin, nil
	}
	return "", errors.New("building a binary needs the Go toolchain, and none was found; run `gperl install toolchain`")
}

// buildConfig is what gperl builds for, as its environment sets it:
//
//	GPERLOS, GPERLARCH  the target platform, by Go's names; this machine's
//	                    when unset
//	GPERLNATIVE         how a binary carries native modules: link (the
//	                    default) links them in statically, load embeds
//	                    them and loads them at run time
//
// Go's own variables play no part: every go command gperl runs has GOOS,
// GOARCH and CGO_ENABLED set from it (goEnv).
type buildConfig struct {
	goos, goarch string
	link         bool
}

// goTarget is a platform as `go tool dist list -json` names it.
type goTarget struct{ GOOS, GOARCH string }

// readBuildConfig reads the build configuration from the environment. A
// target platform the environment names must be one Go builds for.
func readBuildConfig() (buildConfig, error) {
	c := buildConfig{goos: os.Getenv("GPERLOS"), goarch: os.Getenv("GPERLARCH")}
	switch native := os.Getenv("GPERLNATIVE"); native {
	case "", "link":
		c.link = true
	case "load":
	default:
		return c, fmt.Errorf("GPERLNATIVE=%s: want link (link native modules in statically) or load (embed them and load them at run time)", native)
	}
	named := c.goos != "" || c.goarch != ""
	if c.goos == "" {
		c.goos = runtime.GOOS
	}
	if c.goarch == "" {
		c.goarch = runtime.GOARCH
	}
	if !named {
		return c, nil
	}
	goBin, err := goCommand()
	if err != nil {
		return c, fmt.Errorf("GPERLOS=%s GPERLARCH=%s: the platforms Go builds for are the Go toolchain's: %w", c.goos, c.goarch, err)
	}
	cmd := exec.Command(goBin, "tool", "dist", "list", "-json")
	// Not the working directory: a go.mod there could name another Go
	// toolchain to run.
	cmd.Dir = os.TempDir()
	cmd.Env = c.goEnv(false)
	cmd.Stderr = os.Stderr
	out, err := cmd.Output()
	if err != nil {
		return c, fmt.Errorf("go tool dist list: %w", err)
	}
	var targets []goTarget
	if err := json.Unmarshal(out, &targets); err != nil {
		return c, fmt.Errorf("go tool dist list: %w", err)
	}
	if !slices.Contains(targets, goTarget{c.goos, c.goarch}) {
		return c, fmt.Errorf("GPERLOS=%s GPERLARCH=%s: Go does not build for this platform (`go tool dist list` names the ones it does)", c.goos, c.goarch)
	}
	return c, nil
}

// target names the platform c builds for ("linux_arm64").
func (c buildConfig) target() string { return c.goos + "_" + c.goarch }

// goEnv is the environment of a go command gperl runs for c: the
// process's, with Go's target and cgo settings replaced by c's and cgo.
func (c buildConfig) goEnv(cgo bool) []string {
	cgoEnabled := "0"
	if cgo {
		cgoEnabled = "1"
	}
	// The last of duplicate entries is the one a command sees (exec.Cmd).
	return append(os.Environ(), "GOOS="+c.goos, "GOARCH="+c.goarch, "CGO_ENABLED="+cgoEnabled)
}

// muslSuffix ends the name of a Linux target whose C library is not glibc
// (nativeTarget), and muslTag is the build tag a Go build for it carries:
// Go's build constraints tell no C library from another, and a binary for
// one cannot load or link the native modules built for the other.
const (
	muslSuffix = "_musl"
	muslTag    = "musl"
)

// nativeTarget names the platform the native modules for target are built
// for: target, with muslSuffix on Linux when target's C compiler
// (cToolchain) does not use glibc. A target whose compiler cannot be had or
// asked keeps its name: without one only pure-Perl distributions build, and
// a native module's build reports what its compiler lacks.
func nativeTarget(target string) string {
	if goos, _, _ := strings.Cut(target, "_"); goos != "linux" {
		return target
	}
	cc, _, err := cToolchain(target)
	if err != nil {
		return target
	}
	out, err := compilerCommand(strings.Join(cc, " "), "-dM", "-E", "-x", "c", "-include", "features.h", os.DevNull).Output()
	if err != nil {
		fmt.Fprintf(os.Stderr, "gperl: ask the C compiler %s for its C library: %v; native modules are built as for %s\n", strings.Join(cc, " "), err, target)
		return target
	}
	// glibc's features.h defines __GLIBC__ ("#define NAME value" is -dM's
	// format).
	if strings.Contains("\n"+string(out), "\n#define __GLIBC__ ") {
		return target
	}
	return target + muslSuffix
}

// buildFlags is the go build command line for target, up to its output.
func buildFlags(target string) []string {
	if strings.HasSuffix(target, muslSuffix) {
		return []string{"go", "build", "-tags", muslTag}
	}
	return []string{"go", "build"}
}

// targetTags returns the build constraint terms that select target's
// native modules: its GOOS and GOARCH, and on Linux whether muslTag is set.
func targetTags(target string) []string {
	goos, rest, _ := strings.Cut(target, "_")
	goarch, _, _ := strings.Cut(rest, "_")
	tags := []string{goos, goarch}
	switch {
	case goos != "linux":
	case strings.HasSuffix(target, muslSuffix):
		tags = append(tags, muslTag)
	default:
		tags = append(tags, "!"+muslTag)
	}
	return tags
}

// cToolchain returns the C compiler and archiver command lines for
// building native modules for target ("linux_amd64", or with muslSuffix
// for musl): what CC and AR name when set; else the managed Zig, told the
// target when that is another platform; else cc and ar from PATH, for this
// platform.
func cToolchain(target string) (cc, ar []string, err error) {
	platform := strings.TrimSuffix(target, muslSuffix)
	cc, ar = strings.Fields(os.Getenv("CC")), strings.Fields(os.Getenv("AR"))
	zig := managedBin("zig")
	if len(cc) == 0 && zig != "" {
		cc = []string{zig, "cc"}
		if target != hostPlatform() {
			goos, _, _ := strings.Cut(platform, "_")
			abi := map[string]string{"linux": "-gnu"}[goos]
			if platform != target {
				abi = "-musl"
			}
			cc = append(cc, "-target", zigTriple(platform)+abi)
		}
		if len(ar) == 0 {
			ar = []string{zig, "ar"}
		}
	}
	if len(cc) == 0 {
		if platform != hostPlatform() {
			return nil, nil, fmt.Errorf("building native modules for %s needs a C cross compiler: set CC, or run `gperl install toolchain`", target)
		}
		bin, lookErr := exec.LookPath("cc")
		if lookErr != nil {
			return nil, nil, errors.New("building a native (XS) module needs a C compiler, and none was found; run `gperl install toolchain`")
		}
		cc = []string{bin}
	}
	if len(ar) == 0 {
		ar = []string{"ar"}
	}
	return cc, ar, nil
}
