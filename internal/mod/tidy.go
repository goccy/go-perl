package mod

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
)

// CPANfile is the dependency declaration's name in a project directory.
const CPANfile = "cpanfile"

// request is one requirement on its way through resolution.
type request struct {
	Prereq
	pin *Distribution // the distribution the cpanfile's options name, or nil
	by  string        // who requires it, for error messages
}

// provider records which resolved distribution provides a package.
type provider struct {
	version  string
	pathname string
}

// Tidy resolves the project's cpanfile into a complete set of
// distributions, fetches them into the module directory, and writes
// cpanfile.snapshot. Releases the existing snapshot pins stay pinned as
// long as they satisfy the cpanfile; modules named in upgrade are resolved
// afresh.
func (m *Manager) Tidy(ctx context.Context, projectDir string, upgrade ...string) (*Snapshot, error) {
	declared, err := m.helper.CPANfile(filepath.Join(projectDir, CPANfile))
	if err != nil {
		return nil, err
	}
	old, err := ReadSnapshot(filepath.Join(projectDir, SnapshotFile))
	if errors.Is(err, os.ErrNotExist) {
		old, err = &Snapshot{}, nil
	}
	if err != nil {
		return nil, err
	}
	w := &walk{m: m, old: old, fresh: map[string]bool{}, provided: map[string]provider{}, dists: map[string]*Distribution{}, releases: map[string]string{}}
	for _, module := range upgrade {
		w.fresh[module] = true
	}
	var level, named []request // named: those whose options name the distribution
	var conflicts []Prereq
	pins := map[string]request{} // by module
	for _, p := range declared {
		pin, err := pinned(p)
		if err != nil {
			return nil, fmt.Errorf("%s line %d: %s: %w", CPANfile, p.Line, p.Module, err)
		}
		// Optional features are not selected. The project's own tests
		// run too, so its test prereqs count.
		if p.Feature != "" || p.Phase == "develop" {
			continue
		}
		switch p.Type {
		case "requires":
			r := request{Prereq: p, pin: pin, by: CPANfile}
			if other, ok := pins[p.Module]; ok && !samePin(other.pin, pin) {
				return nil, fmt.Errorf("%s lines %d and %d name different distributions for %s", CPANfile, other.Line, p.Line, p.Module)
			}
			pins[p.Module] = r
			if pin != nil {
				named = append(named, r)
			} else {
				level = append(level, r)
			}
		case "conflicts":
			conflicts = append(conflicts, p)
		}
	}
	// The distributions the cpanfile names come first: the other modules
	// they provide meet the requirements on them, which are not resolved
	// elsewhere.
	if len(named) > 0 {
		next, err := w.step(ctx, named)
		if err != nil {
			return nil, err
		}
		level = append(level, next...)
	}
	for len(level) > 0 {
		if level, err = w.step(ctx, level); err != nil {
			return nil, err
		}
	}
	if err := w.conflict(conflicts); err != nil {
		return nil, err
	}
	snap := &Snapshot{}
	for _, d := range w.dists {
		snap.Distributions = append(snap.Distributions, *d)
	}
	sort.Slice(snap.Distributions, func(i, j int) bool { return snap.Distributions[i].Name < snap.Distributions[j].Name })
	return snap, os.WriteFile(filepath.Join(projectDir, SnapshotFile), snap.Format(), 0o644)
}

// walk is the state of one resolution.
type walk struct {
	m        *Manager
	old      *Snapshot
	fresh    map[string]bool          // modules to resolve ignoring the old snapshot
	provided map[string]provider      // package -> resolved provider
	dists    map[string]*Distribution // by tree
	releases map[string]string        // distribution name -> selected tree
}

// step resolves one level of requirements and returns the next: the
// prereqs of the distributions this level added.
func (w *walk) step(ctx context.Context, level []request) ([]request, error) {
	// Requirements already met by the interpreter or by a distribution
	// resolved earlier need no work, unless the cpanfile names the
	// distribution to use.
	var open []request
	seen := map[string]bool{}
	for _, r := range level {
		met := false
		if r.pin == nil {
			var err error
			if met, err = w.met(r); err != nil {
				return nil, err
			}
		}
		if !met && !seen[r.Module+"\x00"+r.Range] {
			seen[r.Module+"\x00"+r.Range] = true
			open = append(open, r)
		}
	}

	// Lookups and downloads are network-bound: run them concurrently.
	picked := make([]*Distribution, len(open))
	// One distribution per source the cpanfile names, however many of its
	// modules it requires: the old snapshot's entry when it still serves
	// every one of them, otherwise one fetch.
	type pinKey struct {
		pathname string
		source   Source
	}
	type pinFetch struct {
		once sync.Once
		d    Distribution
		err  error
		old  *Distribution // the old snapshot's entry while it serves every request so far
		kept bool          // no request has refused old
	}
	fetches := map[pinKey]*pinFetch{}
	for _, r := range open {
		if r.pin == nil {
			continue
		}
		k := pinKey{r.pin.Pathname, r.pin.Source}
		f := fetches[k]
		if f == nil {
			f = &pinFetch{d: *r.pin, kept: true}
			fetches[k] = f
		}
		if !f.kept {
			continue
		}
		old := w.old.Provider(r.Module)
		f.kept = old != nil && !w.fresh[r.Module] && sameSource(r.pin, old) && (f.old == nil || f.old == old)
		if f.kept {
			accepted, err := w.m.helper.Accepts(r.Range, providedVersion(old, r.Module))
			if err != nil {
				return nil, err
			}
			f.kept, f.old = accepted, old
		}
	}
	for _, f := range fetches {
		if f.kept {
			f.d = Distribution{Pathname: f.old.Pathname, Source: f.old.Source}
			f.once.Do(func() {})
		}
	}
	err := each(ctx, indexes(len(open)), func(ctx context.Context, i int) error {
		r := open[i]
		if d := w.old.Provider(r.Module); r.pin == nil && d != nil && !w.fresh[r.Module] && sameSource(nil, d) {
			if ok, err := w.m.helper.Accepts(r.Range, providedVersion(d, r.Module)); err != nil {
				return err
			} else if ok {
				picked[i] = &Distribution{Pathname: d.Pathname, Source: d.Source}
				return nil
			}
		}
		if r.pin != nil {
			// Fetching pins it: its digest or commit becomes known.
			f := fetches[pinKey{r.pin.Pathname, r.pin.Source}]
			// A kept entry's Once has run already, so Do does nothing.
			f.once.Do(func() { _, f.err = w.m.Fetch(ctx, &f.d) })
			if f.err != nil {
				return fmt.Errorf("%s line %d: %s: %w", r.by, r.Line, r.Module, f.err)
			}
			d := f.d
			picked[i] = &d
			return nil
		}
		rel, err := w.m.resolve(ctx, r.Module, r.Range)
		if err != nil {
			return fmt.Errorf("%s requires %s: %w", r.by, r.Module, err)
		}
		picked[i] = &Distribution{Pathname: rel.Pathname}
		return nil
	})
	if err != nil {
		return nil, err
	}
	var added []*Distribution
	keys := make([]string, len(picked))
	for i, d := range picked {
		name, err := distName(d)
		if err != nil {
			return nil, fmt.Errorf("%s requires %s: %w", open[i].by, open[i].Module, err)
		}
		if name == "perl" {
			return nil, fmt.Errorf("%s requires %s, which only perl itself provides and this interpreter does not carry", open[i].by, open[i].Module)
		}
		tree, err := TreePath(w.m.Dir, d)
		if err != nil {
			return nil, fmt.Errorf("%s requires %s: %w", open[i].by, open[i].Module, err)
		}
		keys[i] = tree
		// One release per distribution: two requirements that can only be
		// met by different releases cannot both hold.
		if other, ok := w.releases[name]; ok && other != tree {
			return nil, fmt.Errorf("%s requires %s %s, which needs %s, but %s is already selected",
				open[i].by, open[i].Module, open[i].Range, filepath.Base(tree), filepath.Base(other))
		}
		w.releases[name] = tree
		if _, ok := w.dists[tree]; !ok {
			w.dists[tree] = nil
			added = append(added, d)
		}
	}
	trees := make([]string, len(added))
	err = each(ctx, indexes(len(added)), func(ctx context.Context, i int) error {
		tree, err := w.m.Fetch(ctx, added[i])
		trees[i] = tree
		return err
	})
	if err != nil {
		return nil, err
	}

	// Reading metadata runs on the interpreter, one tree at a time.
	var next []request
	for i, d := range added {
		meta, err := w.m.helper.Meta(trees[i])
		if err != nil {
			return nil, err
		}
		name := strings.TrimSuffix(path.Base(d.Pathname), archiveSuffix(d.Pathname))
		dist := &Distribution{Name: name, Pathname: d.Pathname, Source: d.Source, Provides: meta.Provides}
		reqs := map[string]string{}
		for _, p := range meta.Prereqs {
			// A dependency's own tests are never run.
			if p.Type != "requires" || p.Phase == "test" {
				continue
			}
			if have, ok := reqs[p.Module]; !ok || have == "0" {
				reqs[p.Module] = p.Range
			} else if p.Range != "0" && p.Range != have {
				reqs[p.Module] = have + ", " + p.Range
			}
		}
		for module, rng := range reqs {
			dist.Requirements = append(dist.Requirements, Package{Name: module, Version: rng})
			next = append(next, request{Prereq: Prereq{Module: module, Range: rng}, by: name})
		}
		sort.Slice(next, func(i, j int) bool { return next[i].Module < next[j].Module })
		for _, p := range meta.Provides {
			if _, ok := w.provided[p.Name]; !ok {
				w.provided[p.Name] = provider{version: p.Version, pathname: d.Pathname}
			}
		}
		w.dists[trees[i]] = dist
	}

	// A distribution the cpanfile names must provide the module at a
	// version the requirement accepts; nothing else has checked it.
	for i, r := range open {
		if r.pin == nil {
			continue
		}
		d := w.dists[keys[i]]
		at := slices.IndexFunc(d.Provides, func(p Package) bool { return p.Name == r.Module })
		if at < 0 {
			return nil, fmt.Errorf("%s line %d: %s: %s does not provide it", r.by, r.Line, r.Module, d.Name)
		}
		if ok, err := w.m.helper.Accepts(r.Range, d.Provides[at].Version); err != nil {
			return nil, err
		} else if !ok {
			return nil, fmt.Errorf("%s line %d: %s %s: %s has version %s", r.by, r.Line, r.Module, r.Range, d.Name, d.Provides[at].Version)
		}
	}
	return next, nil
}

// conflict reports a module the cpanfile conflicts with that the
// resolution selected, or that the interpreter carries, at a version the
// conflicts range covers.
func (w *walk) conflict(conflicts []Prereq) error {
	for _, p := range conflicts {
		version, ok := "", false
		if pr, found := w.provided[p.Module]; found {
			version, ok = pr.version, true
		} else {
			var err error
			if version, ok, err = w.m.bundledVersion(p.Module); err != nil {
				return err
			}
		}
		if !ok {
			continue
		}
		hit, err := w.m.helper.Accepts(p.Range, version)
		if err != nil {
			return err
		}
		if hit {
			return fmt.Errorf("%s line %d: conflicts %s %s, but version %s is selected", CPANfile, p.Line, p.Module, p.Range, version)
		}
	}
	return nil
}

// met reports whether a requirement is already satisfied: by the
// interpreter's own modules or by a distribution resolved earlier.
func (w *walk) met(r request) (bool, error) {
	if p, ok := w.provided[r.Module]; ok {
		accepted, err := w.m.helper.Accepts(r.Range, p.version)
		if err != nil {
			return false, err
		}
		if !accepted {
			return false, fmt.Errorf("%s requires %s %s, but %s was resolved to version %s (%s)",
				r.by, r.Module, r.Range, r.Module, p.version, p.pathname)
		}
		return true, nil
	}
	version, ok, err := w.m.bundledVersion(r.Module)
	if err != nil {
		return false, err
	}
	if ok {
		if accepted, err := w.m.helper.Accepts(r.Range, version); err != nil || accepted {
			return accepted, err
		}
	}
	if r.Module == "perl" {
		return false, fmt.Errorf("%s requires perl %s, but this interpreter is perl %s", r.by, r.Range, version)
	}
	return false, nil
}

// bundledVersion caches Helper.Bundled.
func (m *Manager) bundledVersion(module string) (string, bool, error) {
	m.mu.Lock()
	b, ok := m.bundled[module]
	m.mu.Unlock()
	if ok {
		return b.version, b.ok, nil
	}
	version, ok, err := m.helper.Bundled(module)
	if err != nil {
		return "", false, err
	}
	m.mu.Lock()
	m.bundled[module] = bundledModule{version: version, ok: ok}
	m.mu.Unlock()
	return version, ok, nil
}

func providedVersion(d *Distribution, module string) string {
	for _, p := range d.Provides {
		if p.Name == module {
			return p.Version
		}
	}
	return "undef"
}

func archiveSuffix(pathname string) string {
	// The name of a git repository's archive carries a commit, not a
	// version, so distFile need not match it.
	for _, suffix := range []string{".tar.gz", ".tgz", ".tar.bz2", ".zip"} {
		if strings.HasSuffix(pathname, suffix) {
			return suffix
		}
	}
	return path.Ext(pathname)
}

func indexes(n int) []int {
	out := make([]int, n)
	for i := range out {
		out[i] = i
	}
	return out
}

// Download makes every distribution of the project's snapshot present in
// the module directory, resolving the cpanfile first when the project has
// no snapshot yet. A project without a cpanfile has no dependencies.
func (m *Manager) Download(ctx context.Context, projectDir string) (*Snapshot, error) {
	snap, err := ReadSnapshot(filepath.Join(projectDir, SnapshotFile))
	if errors.Is(err, os.ErrNotExist) {
		if _, statErr := os.Stat(filepath.Join(projectDir, CPANfile)); errors.Is(statErr, os.ErrNotExist) {
			return &Snapshot{}, nil
		}
		return m.Tidy(ctx, projectDir)
	}
	if err != nil {
		return nil, err
	}
	err = each(ctx, snap.Distributions, func(ctx context.Context, d Distribution) error {
		_, err := m.Fetch(ctx, &d)
		return err
	})
	return snap, err
}

// moduleSpec matches "Module::Name" and "Module::Name@1.23".
var moduleSpec = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*(?:::[A-Za-z0-9_]+)*)(?:@(.+))?$`)

// Get adds the modules to the project's cpanfile — "Module" for the latest
// release, "Module@1.23" for exactly that version — replacing a module's
// existing requires line, and resolves again.
func (m *Manager) Get(ctx context.Context, projectDir string, specs ...string) (*Snapshot, error) {
	file := filepath.Join(projectDir, CPANfile)
	data, err := os.ReadFile(file)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, err
	}
	text := string(data)
	var modules []string
	for _, spec := range specs {
		match := moduleSpec.FindStringSubmatch(spec)
		if match == nil {
			return nil, fmt.Errorf("%q: want Module::Name or Module::Name@version", spec)
		}
		line := fmt.Sprintf("requires '%s';", match[1])
		if match[2] != "" {
			line = fmt.Sprintf("requires '%s', '== %s';", match[1], match[2])
		}
		existing := regexp.MustCompile(`(?m)^requires\s+['"]` + regexp.QuoteMeta(match[1]) + `['"].*$`)
		if existing.MatchString(text) {
			text = existing.ReplaceAllLiteralString(text, line)
		} else {
			if text != "" && !strings.HasSuffix(text, "\n") {
				text += "\n"
			}
			text += line + "\n"
		}
		modules = append(modules, match[1])
	}
	if err := os.WriteFile(file, []byte(text), 0o644); err != nil {
		return nil, err
	}
	return m.Tidy(ctx, projectDir, modules...)
}

// Meta reads the metadata of the distribution tree at dir.
func (m *Manager) Meta(dir string) (*Meta, error) { return m.helper.Meta(dir) }

// Clean removes the module directory.
func (m *Manager) Clean() error { return os.RemoveAll(m.Dir) }
