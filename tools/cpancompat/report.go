package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

// coverage counts, for one set of distributions, how a run went.
type coverage struct {
	// Distributions is the size of the set.
	Distributions int `json:"distributions"`
	// Recorded is how many of them the run has a result for.
	Recorded int `json:"recorded"`
	// Outcomes counts the recorded ones by outcome.
	Outcomes map[string]int `json:"outcomes"`
	// PassShareOfRecorded and PassShareOfAll are the share of passing
	// distributions among the recorded ones and among the whole set.
	PassShareOfRecorded float64 `json:"pass_share_of_recorded"`
	PassShareOfAll      float64 `json:"pass_share_of_all"`
}

// blocker is a distribution that others were not attempted because of.
type blocker struct {
	Dist    string `json:"dist"`
	Outcome string `json:"outcome"`
	// Blocked counts the recorded distributions that name it as what
	// blocked them; DependentsTotal is everything that needs it.
	Blocked         int `json:"blocked"`
	DependentsTotal int `json:"dependents_total"`
}

// runSummary is what report writes to summary.json.
type runSummary struct {
	Run string `json:"run"`
	// All is every distribution of the snapshot. PassOnPerl are the ones
	// CPAN Testers saw pass more often than fail on perl itself, the
	// reference this survey measures against; AnyPassOnPerl the ones with
	// at least one passing report.
	All           coverage  `json:"all"`
	PassOnPerl    coverage  `json:"pass_on_perl"`
	AnyPassOnPerl coverage  `json:"any_pass_on_perl"`
	Blockers      []blocker `json:"blockers"`
}

// report summarizes a run: outcomes, coverage against the distributions
// that pass on perl, and what blocks the most.
func report(dataDir, run string, w io.Writer) error {
	dists, err := readDists(filepath.Join(dataDir, "dists.jsonl"))
	if err != nil {
		return err
	}
	rows, err := readJSONLines[result](filepath.Join(dataDir, "runs", run, "results.jsonl"))
	if err != nil {
		return err
	}
	// A distribution recorded again (a later run with -redo) counts by its
	// last record.
	latest := map[string]*result{}
	for i := range rows {
		latest[rows[i].Dist] = &rows[i]
	}

	s := runSummary{Run: run}
	sets := []struct {
		c      *coverage
		member func(d *dist) bool
	}{
		{&s.All, func(*dist) bool { return true }},
		{&s.PassOnPerl, func(d *dist) bool { return d.Testers != nil && d.Testers.Pass > d.Testers.Fail }},
		{&s.AnyPassOnPerl, func(d *dist) bool { return d.Testers != nil && d.Testers.Pass > 0 }},
	}
	for _, set := range sets {
		set.c.Outcomes = map[string]int{}
	}
	dependents := map[string]int{}
	for i := range dists {
		d := &dists[i]
		dependents[d.Dist] = d.DependentsTotal
		for _, set := range sets {
			if !set.member(d) {
				continue
			}
			set.c.Distributions++
			if res := latest[d.Dist]; res != nil {
				set.c.Recorded++
				set.c.Outcomes[res.Outcome]++
			}
		}
	}
	for _, set := range sets {
		c := set.c
		if c.Recorded > 0 {
			c.PassShareOfRecorded = float64(c.Outcomes[outcomePass]) / float64(c.Recorded)
		}
		if c.Distributions > 0 {
			c.PassShareOfAll = float64(c.Outcomes[outcomePass]) / float64(c.Distributions)
		}
	}

	blocked := map[string]int{}
	for _, res := range latest {
		for _, b := range res.BlockedBy {
			blocked[b]++
		}
	}
	for name, n := range blocked {
		b := blocker{Dist: name, Blocked: n, DependentsTotal: dependents[name]}
		if res := latest[name]; res != nil {
			b.Outcome = res.Outcome
		}
		s.Blockers = append(s.Blockers, b)
	}
	sort.Slice(s.Blockers, func(i, j int) bool {
		a, b := s.Blockers[i], s.Blockers[j]
		if a.DependentsTotal != b.DependentsTotal {
			return a.DependentsTotal > b.DependentsTotal
		}
		return a.Dist < b.Dist
	})

	raw, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(dataDir, "runs", run, "summary.json"), append(raw, '\n'), 0o644); err != nil {
		return err
	}

	print := func(title string, c coverage) {
		fmt.Fprintf(w, "%s: %d distributions, %d recorded, %d pass (%.1f%% of recorded, %.1f%% of all)\n",
			title, c.Distributions, c.Recorded, c.Outcomes[outcomePass], 100*c.PassShareOfRecorded, 100*c.PassShareOfAll)
		names := make([]string, 0, len(c.Outcomes))
		for name := range c.Outcomes {
			names = append(names, name)
		}
		sort.Slice(names, func(i, j int) bool {
			if c.Outcomes[names[i]] != c.Outcomes[names[j]] {
				return c.Outcomes[names[i]] > c.Outcomes[names[j]]
			}
			return names[i] < names[j]
		})
		for _, name := range names {
			fmt.Fprintf(w, "  %-22s %6d\n", name, c.Outcomes[name])
		}
	}
	print("all of CPAN", s.All)
	print("passing on perl (more passing than failing CPAN Testers reports)", s.PassOnPerl)
	print("with at least one passing CPAN Testers report", s.AnyPassOnPerl)
	if len(s.Blockers) > 0 {
		fmt.Fprintln(w, "not installable, by how much depends on them:")
		for i, b := range s.Blockers {
			if i == 30 {
				break
			}
			fmt.Fprintf(w, "  %-32s %-20s needed by %5d, blocked %4d here\n", b.Dist, b.Outcome, b.DependentsTotal, b.Blocked)
		}
	}
	return nil
}
