package mod

import (
	"context"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

//go:embed helper.pl
var helperProgram string

// Helper answers the questions that are Perl's own semantics (cpanfile
// evaluation, version ranges, distribution metadata) on an interpreter
// instance, started on first use. It is safe for concurrent use.
type Helper struct {
	start sync.Once
	err   error // why the interpreter did not start
	mu    sync.Mutex
	p     *perl.Perl
}

// boot starts the interpreter and loads the helper program.
func (h *Helper) boot() {
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		h.err = fmt.Errorf("extract embedded stdlib: %w", err)
		return
	}
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, Stderr: os.Stderr})
	if err != nil {
		h.err = err
		return
	}
	r, err := p.Eval(context.Background(), helperProgram)
	if err == nil && r.Error != nil {
		err = r.Error
	}
	if err != nil {
		h.err = errors.Join(fmt.Errorf("load the dependency helper: %w", err), p.Close())
		return
	}
	h.p = p
}

// Close stops the interpreter, if it was started.
func (h *Helper) Close() error {
	h.start.Do(func() { h.err = errors.New("dependency helper is closed") })
	if h.p == nil {
		return nil
	}
	return h.p.Close()
}

// call runs one helper sub and returns its list as strings.
func (h *Helper) call(name string, args ...string) ([]string, error) {
	h.start.Do(h.boot)
	if h.err != nil {
		return nil, h.err
	}
	vals := make([]perl.Value, len(args))
	for i, a := range args {
		vals[i] = perl.ValueOf(a)
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	res, err := h.p.Call(context.Background(), "GperlMod::"+name, vals...)
	if err != nil {
		return nil, err
	}
	out := make([]string, len(res))
	for i, v := range res {
		s, err := perl.As[perl.ScalarValue](v)
		if err != nil {
			return nil, fmt.Errorf("GperlMod::%s: result %d: %w", name, i, err)
		}
		out[i] = s.String()
	}
	return out, nil
}

// Prereq is one dependency declaration.
type Prereq struct {
	Module string
	Range  string // a CPAN::Meta version range; "0" accepts any version
	Phase  string // configure, build, test, runtime, develop
	Type   string // requires, recommends, suggests, conflicts
}

func prereqsOf(fields []string) ([]Prereq, error) {
	if len(fields)%4 != 0 {
		return nil, fmt.Errorf("malformed prereq list (%d fields)", len(fields))
	}
	out := make([]Prereq, 0, len(fields)/4)
	for i := 0; i < len(fields); i += 4 {
		out = append(out, Prereq{Module: fields[i], Range: fields[i+1], Phase: fields[i+2], Type: fields[i+3]})
	}
	return out, nil
}

// CPANfile evaluates the cpanfile at path and returns what it declares.
func (h *Helper) CPANfile(path string) ([]Prereq, error) {
	// The interpreter's working directory is not the process's.
	path, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	fields, err := h.call("cpanfile", path)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	return prereqsOf(fields)
}

// Best returns the index of the highest of versions that rng accepts, or
// -1 when none does.
func (h *Helper) Best(rng string, versions ...string) (int, error) {
	res, err := h.call("best", append([]string{rng}, versions...)...)
	if err != nil {
		return 0, fmt.Errorf("version range %q: %w", rng, err)
	}
	if len(res) != 1 {
		return 0, fmt.Errorf("version range %q: helper returned %d values", rng, len(res))
	}
	return strconv.Atoi(res[0])
}

// Accepts reports whether rng accepts version.
func (h *Helper) Accepts(rng, version string) (bool, error) {
	i, err := h.Best(rng, version)
	return i == 0, err
}

// Bundled returns the version of module that ships with the interpreter,
// and whether it ships at all. "perl" names the interpreter itself.
func (h *Helper) Bundled(module string) (version string, ok bool, err error) {
	res, err := h.call("bundled", module)
	if err != nil {
		return "", false, err
	}
	if len(res) != 1 {
		return "", false, fmt.Errorf("bundled %s: helper returned %d values", module, len(res))
	}
	return res[0], res[0] != "", nil
}

// Meta is what a distribution tree says about itself.
type Meta struct {
	// DynamicConfig: the prereqs are only final after the distribution's
	// configure step has run.
	DynamicConfig bool
	// StaticInstall: the distribution installs by copying lib/ (META
	// x_static_install), with no configure or build step.
	StaticInstall bool
	// HasXS: the tree carries XS or C sources.
	HasXS    bool
	Prereqs  []Prereq
	Provides []Package
}

// Meta reads the metadata of the distribution extracted at dir.
func (h *Helper) Meta(dir string) (*Meta, error) {
	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	fields, err := h.call("meta", dir)
	if err != nil {
		return nil, fmt.Errorf("%s: read metadata: %w", dir, err)
	}
	m := &Meta{}
	for i := 0; i < len(fields); {
		if fields[i] != "-" || i+1 >= len(fields) {
			return nil, fmt.Errorf("%s: malformed metadata record at field %d", dir, i)
		}
		key, rest := fields[i+1], fields[i+2:]
		width := map[string]int{"dynamic_config": 1, "static_install": 1, "has_xs": 1, "prereq": 4, "provides": 2}[key]
		if width == 0 || len(rest) < width {
			return nil, fmt.Errorf("%s: malformed metadata record %q", dir, key)
		}
		switch key {
		case "dynamic_config":
			m.DynamicConfig = rest[0] == "1"
		case "static_install":
			m.StaticInstall = rest[0] == "1"
		case "has_xs":
			m.HasXS = rest[0] == "1"
		case "prereq":
			m.Prereqs = append(m.Prereqs, Prereq{Module: rest[0], Range: rest[1], Phase: rest[2], Type: rest[3]})
		case "provides":
			m.Provides = append(m.Provides, Package{Name: rest[0], Version: rest[1]})
		}
		i += 2 + width
	}
	return m, nil
}
