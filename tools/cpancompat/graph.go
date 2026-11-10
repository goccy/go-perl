package main

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	distsFile = "dists.jsonl"
	// providedFile lists the modules the interpreter under survey ships.
	providedFile = "provided.json"
	summaryFile  = "graph-summary.json"

	// perlDist is the distribution name of the interpreter itself. A
	// module the package index assigns to it ships with perl and is never
	// installed from CPAN.
	perlDist = "perl"
)

// release is one MetaCPAN release document (the fields fetch keeps).
type release struct {
	Distribution string       `json:"distribution"`
	Name         string       `json:"name"`
	Version      versionText  `json:"version"`
	Author       string       `json:"author"`
	Archive      string       `json:"archive"`
	DownloadURL  string       `json:"download_url"`
	SHA256       string       `json:"checksum_sha256"`
	Date         string       `json:"date"`
	Maturity     string       `json:"maturity"`
	Authorized   bool         `json:"authorized"`
	Dependency   []dependency `json:"dependency"`
	Tests        *testers     `json:"tests"`
	Stat         struct {
		Size count `json:"size"`
	} `json:"stat"`
}

type dependency struct {
	Module       string      `json:"module"`
	Version      versionText `json:"version"`
	Phase        string      `json:"phase"`
	Relationship string      `json:"relationship"`
}

// testers is the CPAN Testers report summary of one release.
type testers struct {
	Pass    count `json:"pass"`
	Fail    count `json:"fail"`
	NA      count `json:"na"`
	Unknown count `json:"unknown"`
}

// requirement is one declared dependency with the distribution that
// provides the module according to the package index. Dist is "" when the
// index does not know the module.
type requirement struct {
	Module       string `json:"module"`
	Version      string `json:"version,omitempty"`
	Phase        string `json:"phase"`
	Relationship string `json:"relationship"`
	Dist         string `json:"dist,omitempty"`
	// Provided is set when the interpreter's own standard library ships
	// the module at a version that meets the requirement, so nothing has
	// to be installed for it.
	Provided bool `json:"provided,omitempty"`
}

// dist is the record the later steps work from: one per distribution.
type dist struct {
	Dist    string `json:"dist"`
	Release string `json:"release"`
	Version string `json:"version"`
	Author  string `json:"author"`
	// Path is the archive location below authors/id/ on any CPAN mirror.
	Path    string   `json:"path"`
	SHA256  string   `json:"sha256,omitempty"`
	Size    int64    `json:"size,omitempty"`
	Date    string   `json:"date,omitempty"`
	Modules []string `json:"modules"`

	Requires []requirement `json:"requires,omitempty"`
	// Needs is the set of other distributions that must be installed to
	// configure, build, test and run this one (relationship "requires").
	Needs []string `json:"needs,omitempty"`
	// PerlRequirement is the declared minimum perl version, if any.
	PerlRequirement string `json:"perl_requirement,omitempty"`
	// Unresolved lists required modules the package index does not know.
	Unresolved []string `json:"unresolved,omitempty"`

	Testers *testers `json:"testers,omitempty"`

	// Level is the length of the longest dependency chain below this
	// distribution: every distribution only needs lower levels (members
	// of one dependency cycle share a level).
	Level int `json:"level"`
	// Cycle numbers the dependency cycle this distribution is part of
	// (0 = none).
	Cycle int `json:"cycle,omitempty"`
	// NeedsTotal counts the distributions needed directly or indirectly.
	NeedsTotal int `json:"needs_total"`
	// Dependents and DependentsTotal count the distributions that need
	// this one directly, and directly or indirectly.
	Dependents      int `json:"dependents"`
	DependentsTotal int `json:"dependents_total"`
}

type graphSummary struct {
	Distributions       int            `json:"distributions"`
	DuplicateLatest     int            `json:"extra_latest_releases_dropped"`
	Modules             int            `json:"modules"`
	IndexedModules      int            `json:"indexed_modules"`
	NoIndexedModules    int            `json:"distributions_without_indexed_modules"`
	IndexOnlyArchives   int            `json:"index_archives_of_unknown_release"`
	Edges               int            `json:"edges"`
	NoNeeds             int            `json:"distributions_needing_nothing"`
	WithUnresolved      int            `json:"distributions_with_unresolved_requirements"`
	ProvidedModules     int            `json:"modules_provided_by_interpreter"`
	UnreadableVersions  int            `json:"requirements_with_unreadable_version"`
	Cycles              int            `json:"cycles"`
	InCycles            int            `json:"distributions_in_cycles"`
	MaxLevel            int            `json:"max_level"`
	ByLevel             map[int]int    `json:"distributions_by_level"`
	TestersWithReports  int            `json:"distributions_with_tester_reports"`
	TestersAnyPass      int            `json:"distributions_with_a_passing_report"`
	TestersMajorityPass int            `json:"distributions_with_more_pass_than_fail"`
	MostDependedOn      []dependedUpon `json:"most_depended_on"`
}

type dependedUpon struct {
	Dist            string `json:"dist"`
	DependentsTotal int    `json:"dependents_total"`
}

// requiredPhases are the phases whose "requires" must be installed before
// a distribution's tests can run.
var requiredPhases = map[string]bool{"configure": true, "build": true, "test": true, "runtime": true}

func buildGraph(dataDir string) error {
	moduleArchive, err := readPackages(filepath.Join(dataDir, packagesFile))
	if err != nil {
		return err
	}
	provided, err := readProvided(filepath.Join(dataDir, providedFile))
	if err != nil {
		return err
	}
	releases, err := readReleases(filepath.Join(dataDir, releasesFile))
	if err != nil {
		return err
	}

	// Join the two sources on the archive path below authors/id/.
	indexedInArchive := map[string]int{}
	for _, path := range moduleArchive {
		indexedInArchive[path]++
	}
	// MetaCPAN marks more than one release of a few distributions latest.
	// The one the package index points more modules at is the one CPAN
	// clients install; the newer upload breaks a tie.
	chosen := map[string]int{}
	paths := make([]string, len(releases))
	duplicates := 0
	for i, r := range releases {
		_, path, ok := strings.Cut(r.DownloadURL, "/authors/id/")
		if !ok {
			return fmt.Errorf("release %s: download URL %q has no authors/id/ path", r.Name, r.DownloadURL)
		}
		paths[i] = path
		j, dup := chosen[r.Distribution]
		if !dup {
			chosen[r.Distribution] = i
			continue
		}
		duplicates++
		a, b := indexedInArchive[path], indexedInArchive[paths[j]]
		if a > b || (a == b && r.Date > releases[j].Date) {
			chosen[r.Distribution] = i
		}
	}

	archiveDist := map[string]string{}
	dists := make([]*dist, 0, len(chosen))
	byName := map[string]*dist{}
	releaseByDist := map[string]*release{}
	for i := range releases {
		r, path := &releases[i], paths[i]
		if chosen[r.Distribution] != i {
			continue
		}
		releaseByDist[r.Distribution] = r
		d := &dist{
			Dist: r.Distribution, Release: r.Name, Version: string(r.Version), Author: r.Author,
			Path: path, SHA256: r.SHA256, Size: int64(r.Stat.Size), Date: r.Date, Testers: r.Tests,
			Modules: []string{},
		}
		dists = append(dists, d)
		byName[d.Dist] = d
		archiveDist[path] = d.Dist
	}
	sort.Slice(dists, func(i, j int) bool { return dists[i].Dist < dists[j].Dist })

	// Archives the index still points at although they are not a latest
	// release: they name the distribution a module belongs to, without
	// becoming a record of their own.
	indexedOnly, err := readReleases(filepath.Join(dataDir, indexedOnlyFile))
	if err != nil {
		return err
	}
	for _, r := range indexedOnly {
		// The archive-name query also returns same-named uploads that
		// were deleted from CPAN (no download URL); the index cannot be
		// pointing at those.
		_, path, ok := strings.Cut(r.DownloadURL, "/authors/id/")
		if !ok {
			continue
		}
		if _, isLatest := archiveDist[path]; !isLatest && indexedInArchive[path] > 0 {
			archiveDist[path] = r.Distribution
		}
	}

	sum := graphSummary{
		Distributions: len(dists), Modules: len(moduleArchive),
		DuplicateLatest: duplicates, ByLevel: map[int]int{},
		ProvidedModules: len(provided),
	}
	indexOnly := map[string]bool{}
	for module, path := range moduleArchive {
		name, ok := archiveDist[path]
		if !ok {
			indexOnly[path] = true
			continue
		}
		if d := byName[name]; d != nil && d.Path == path {
			d.Modules = append(d.Modules, module)
			sum.IndexedModules++
		}
	}
	sum.IndexOnlyArchives = len(indexOnly)

	index := map[string]int{}
	for i, d := range dists {
		index[d.Dist] = i
		sort.Strings(d.Modules)
		if len(d.Modules) == 0 {
			sum.NoIndexedModules++
		}
	}

	// Resolve every declared dependency to the distribution providing it.
	needs := make([][]int, len(dists))
	for i, d := range dists {
		seen := map[int]bool{}
		unresolved := map[string]bool{}
		for _, dep := range releaseByDist[d.Dist].Dependency {
			req := requirement{Module: dep.Module, Version: string(dep.Version), Phase: dep.Phase, Relationship: dep.Relationship}
			if req.Version == "0" {
				req.Version = ""
			}
			required := dep.Relationship == "requires" && requiredPhases[dep.Phase]
			if dep.Module == "perl" {
				req.Dist = perlDist
				if required && dep.Phase == "runtime" {
					d.PerlRequirement = string(dep.Version)
				}
				d.Requires = append(d.Requires, req)
				continue
			}
			if path, ok := moduleArchive[dep.Module]; ok {
				req.Dist = archiveDist[path]
			}
			if have, shipped := provided[dep.Module]; shipped {
				satisfied, ok := satisfies(have, string(dep.Version))
				if !ok {
					sum.UnreadableVersions++
				}
				req.Provided = satisfied
			}
			d.Requires = append(d.Requires, req)
			if !required || req.Provided {
				continue
			}
			if req.Dist == "" {
				unresolved[dep.Module] = true
				continue
			}
			if req.Dist == perlDist {
				continue
			}
			// A module owned by a superseded upload of a distribution
			// that has no latest release cannot be scheduled.
			j, ok := index[req.Dist]
			if !ok {
				unresolved[dep.Module] = true
				continue
			}
			if j != i && !seen[j] {
				seen[j] = true
				needs[i] = append(needs[i], j)
			}
		}
		sort.Slice(d.Requires, func(a, b int) bool {
			x, y := d.Requires[a], d.Requires[b]
			if x.Phase != y.Phase {
				return x.Phase < y.Phase
			}
			if x.Relationship != y.Relationship {
				return x.Relationship < y.Relationship
			}
			return x.Module < y.Module
		})
		sort.Ints(needs[i])
		for _, j := range needs[i] {
			d.Needs = append(d.Needs, dists[j].Dist)
		}
		for m := range unresolved {
			d.Unresolved = append(d.Unresolved, m)
		}
		sort.Strings(d.Unresolved)
		sum.Edges += len(needs[i])
		if len(needs[i]) == 0 {
			sum.NoNeeds++
		}
		if len(d.Unresolved) > 0 {
			sum.WithUnresolved++
		}
	}

	dependents := make([][]int, len(dists))
	for i, ns := range needs {
		for _, j := range ns {
			dependents[j] = append(dependents[j], i)
		}
	}

	// Cycles and levels.
	component, components := stronglyConnected(needs)
	cycleNumber := map[int]int{}
	for c, members := range components {
		if len(members) > 1 {
			sum.Cycles++
			sum.InCycles += len(members)
			cycleNumber[c] = sum.Cycles
		}
	}
	// stronglyConnected emits components dependencies-first, so one pass
	// in that order sees every needed component's level before its users.
	componentLevel := make([]int, len(components))
	for c, members := range components {
		level := 0
		for _, i := range members {
			for _, j := range needs[i] {
				if cj := component[j]; cj != c && componentLevel[cj]+1 > level {
					level = componentLevel[cj] + 1
				}
			}
		}
		componentLevel[c] = level
	}

	mark := make([]int, len(dists))
	for i := range mark {
		mark[i] = -1
	}
	reach := func(start int, adj [][]int) int {
		n, stack := 0, []int{start}
		mark[start] = start
		for len(stack) > 0 {
			v := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			for _, w := range adj[v] {
				if mark[w] != start {
					mark[w] = start
					n++
					stack = append(stack, w)
				}
			}
		}
		return n
	}
	for i, d := range dists {
		d.Level = componentLevel[component[i]]
		d.Cycle = cycleNumber[component[i]]
		d.Dependents = len(dependents[i])
		d.NeedsTotal = reach(i, needs)
	}
	for i := range mark {
		mark[i] = -1
	}
	for i, d := range dists {
		d.DependentsTotal = reach(i, dependents)
		sum.ByLevel[d.Level]++
		if d.Level > sum.MaxLevel {
			sum.MaxLevel = d.Level
		}
		if t := d.Testers; t != nil && t.Pass+t.Fail+t.NA+t.Unknown > 0 {
			sum.TestersWithReports++
			if t.Pass > 0 {
				sum.TestersAnyPass++
			}
			if t.Pass > t.Fail {
				sum.TestersMajorityPass++
			}
		}
	}

	top := append([]*dist(nil), dists...)
	sort.SliceStable(top, func(i, j int) bool { return top[i].DependentsTotal > top[j].DependentsTotal })
	for _, d := range top[:min(30, len(top))] {
		sum.MostDependedOn = append(sum.MostDependedOn, dependedUpon{d.Dist, d.DependentsTotal})
	}

	if err := writeJSONLines(filepath.Join(dataDir, distsFile), dists); err != nil {
		return err
	}
	out, err := json.MarshalIndent(sum, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dataDir, summaryFile), append(out, '\n'), 0o644); err != nil {
		return err
	}
	os.Stdout.Write(append(out, '\n'))
	return nil
}

// readPackages parses the CPAN package index into module -> archive path.
// The format is PAUSE's: mail-style header lines, a blank line, then one
// "module version path" row per indexed module.
func readPackages(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	zr, err := gzip.NewReader(f)
	if err != nil {
		return nil, err
	}
	modules := map[string]string{}
	sc := bufio.NewScanner(zr)
	inBody := false
	for line := 1; sc.Scan(); line++ {
		if !inBody {
			inBody = sc.Text() == ""
			continue
		}
		fields := strings.Fields(sc.Text())
		if len(fields) != 3 {
			return nil, fmt.Errorf("%s: line %d: want 3 columns, got %d", path, line, len(fields))
		}
		modules[fields[0]] = fields[2]
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(modules) == 0 {
		return nil, fmt.Errorf("%s: no module rows", path)
	}
	return modules, nil
}

// readProvided reads the module -> version map of the interpreter's
// standard library (the output of provided.pl run on that interpreter).
// A module without a declared version maps to "".
func readProvided(path string) (map[string]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w (generate it by running provided.pl on the interpreter under survey)", err)
	}
	var versions map[string]*string
	if err := json.Unmarshal(raw, &versions); err != nil {
		return nil, fmt.Errorf("%s: %w", path, err)
	}
	provided := make(map[string]string, len(versions))
	for module, v := range versions {
		if v != nil {
			provided[module] = *v
		} else {
			provided[module] = ""
		}
	}
	return provided, nil
}

func readReleases(path string) ([]release, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var releases []release
	dec := json.NewDecoder(bufio.NewReaderSize(f, 1<<20))
	for dec.More() {
		var r release
		if err := dec.Decode(&r); err != nil {
			return nil, fmt.Errorf("%s: %w", path, err)
		}
		releases = append(releases, r)
	}
	return releases, nil
}

func writeJSONLines[T any](path string, rows []T) error {
	f, err := os.Create(path + ".tmp")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	defer f.Close()
	w := bufio.NewWriterSize(f, 1<<20)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	for _, row := range rows {
		if err := enc.Encode(row); err != nil {
			return err
		}
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

// stronglyConnected returns, for the graph adj, each node's component and
// the components' members. Components come out dependencies-first: every
// component appears after all components its members point to. Iterative
// Tarjan, because dependency chains are too deep for recursion to be a
// safe assumption.
func stronglyConnected(adj [][]int) (component []int, components [][]int) {
	n := len(adj)
	const unvisited = -1
	order := make([]int, n)
	low := make([]int, n)
	onStack := make([]bool, n)
	component = make([]int, n)
	for i := range order {
		order[i] = unvisited
	}
	type frame struct{ node, next int }
	var stack []int
	var calls []frame
	counter := 0
	for root := 0; root < n; root++ {
		if order[root] != unvisited {
			continue
		}
		calls = append(calls, frame{root, 0})
		for len(calls) > 0 {
			fr := &calls[len(calls)-1]
			v := fr.node
			if fr.next == 0 {
				order[v], low[v] = counter, counter
				counter++
				stack = append(stack, v)
				onStack[v] = true
			}
			if fr.next < len(adj[v]) {
				w := adj[v][fr.next]
				fr.next++
				if order[w] == unvisited {
					calls = append(calls, frame{w, 0})
				} else if onStack[w] && order[w] < low[v] {
					low[v] = order[w]
				}
				continue
			}
			if low[v] == order[v] {
				var members []int
				for {
					w := stack[len(stack)-1]
					stack = stack[:len(stack)-1]
					onStack[w] = false
					component[w] = len(components)
					members = append(members, w)
					if w == v {
						break
					}
				}
				components = append(components, members)
			}
			calls = calls[:len(calls)-1]
			if len(calls) > 0 {
				if p := calls[len(calls)-1].node; low[v] < low[p] {
					low[p] = low[v]
				}
			}
		}
	}
	return component, components
}
