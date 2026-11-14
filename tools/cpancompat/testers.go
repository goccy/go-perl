package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// testersAPI is the CPAN Testers API. Unlike the per-release totals
// MetaCPAN carries, its report summaries name the perl version, operating
// system and grade of every single report.
const testersAPI = "https://api.cpantesters.org/v3"

// testerGrades are the grades a CPAN Testers report can have.
var testerGrades = []string{"pass", "fail", "na", "unknown"}

// testerReport is one CPAN Testers report summary, as the API gives it.
type testerReport struct {
	GUID     string `json:"guid"`
	Dist     string `json:"dist"`
	Version  string `json:"version"`
	Grade    string `json:"grade"`
	Perl     string `json:"perl"`
	OSName   string `json:"osname"`
	Platform string `json:"platform"`
	Date     string `json:"date"`
}

// testersFile is the file of the reports made on one perl version.
func testersFile(perl string) string {
	return "testers-perl-" + perl + ".jsonl"
}

// testersClient has no overall time limit: the answer for one perl version
// and grade is every such report ever made, a long stream.
var testersClient = &http.Client{}

// fetchTesters downloads every CPAN Testers report made on the given perl
// version (as CPAN Testers names it, e.g. 5.44.0), of every grade, into
// testers-perl-<version>.jsonl, one report per line.
func fetchTesters(dataDir, perl string) error {
	if perl == "" {
		return fmt.Errorf("testers: no perl version (-perl)")
	}
	path := filepath.Join(dataDir, testersFile(perl))
	tmp, err := os.Create(path + ".tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	defer tmp.Close()
	w := bufio.NewWriter(tmp)
	seen := map[string]bool{}
	for _, grade := range testerGrades {
		n := 0
		u := testersAPI + "/summary?" + url.Values{"perl": {perl}, "grade": {grade}}.Encode()
		var lastErr error
		for attempt := 0; attempt < 5; attempt++ {
			if attempt > 0 {
				time.Sleep(time.Duration(attempt) * 10 * time.Second)
				fmt.Fprintf(os.Stderr, "testers: %s: again after %v\n", grade, lastErr)
			}
			// A stream that breaks off is asked for again from the start;
			// reports already written are recognized by their id.
			lastErr = streamReports(u, func(r *testerReport) error {
				if seen[r.GUID] {
					return nil
				}
				seen[r.GUID] = true
				n++
				line, err := json.Marshal(r)
				if err != nil {
					return err
				}
				_, err = w.Write(append(line, '\n'))
				return err
			})
			if lastErr == nil {
				break
			}
		}
		if lastErr != nil {
			return fmt.Errorf("testers: %s reports on perl %s: %w", grade, perl, lastErr)
		}
		fmt.Fprintf(os.Stderr, "testers: perl %s: %d %s reports\n", perl, n, grade)
	}
	if err := w.Flush(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

// streamReports requests u, whose answer is a JSON array of report
// summaries, and hands each element to fn as it is decoded.
func streamReports(u string, fn func(*testerReport) error) error {
	req, err := http.NewRequest(http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "cpancompat (go-perl CPAN compatibility survey)")
	resp, err := testersClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return fmt.Errorf("%s: %s: %s", u, resp.Status, strings.TrimSpace(string(body)))
	}
	dec := json.NewDecoder(bufio.NewReaderSize(resp.Body, 1<<20))
	if tok, err := dec.Token(); err != nil {
		return err
	} else if delim, ok := tok.(json.Delim); !ok || delim != '[' {
		return fmt.Errorf("%s: the answer is not an array", u)
	}
	for dec.More() {
		var r testerReport
		if err := dec.Decode(&r); err != nil {
			return err
		}
		if err := fn(&r); err != nil {
			return err
		}
	}
	_, err = dec.Token() // the closing bracket: the array arrived whole
	return err
}

// perlReports counts the CPAN Testers reports one release got on one perl
// version: on every operating system, and on Linux, the system the survey
// runs on.
type perlReports struct {
	Perl      string `json:"perl"`
	Pass      int    `json:"pass"`
	Fail      int    `json:"fail"`
	NA        int    `json:"na"`
	Unknown   int    `json:"unknown"`
	LinuxPass int    `json:"linux_pass"`
	LinuxFail int    `json:"linux_fail"`
}

// passesOnPerl reports whether d belongs to the reference set the survey
// measures against: the release's CPAN Testers reports on the perl version
// the interpreter is built from pass more often than they fail on Linux,
// the system the survey runs on. A release nobody tested on that perl is
// outside the set.
func (d *dist) passesOnPerl() bool {
	p := d.OnPerl
	return p != nil && p.LinuxPass > p.LinuxFail
}

// graphPerl returns the perl version whose CPAN Testers reports the graph
// of dataDir carries ("" when it was built without -perl).
func graphPerl(dataDir string) (string, error) {
	raw, err := os.ReadFile(filepath.Join(dataDir, summaryFile))
	if err != nil {
		return "", err
	}
	var sum struct {
		Perl string `json:"perl"`
	}
	if err := json.Unmarshal(raw, &sum); err != nil {
		return "", fmt.Errorf("%s: %w", summaryFile, err)
	}
	return sum.Perl, nil
}

// readTesters reads testers-perl-<version>.jsonl into counts per release
// name (distribution and version joined by "-", the way CPAN names the
// release, which is how MetaCPAN names it too).
func readTesters(dataDir, perl string) (map[string]*perlReports, error) {
	f, err := os.Open(filepath.Join(dataDir, testersFile(perl)))
	if err != nil {
		return nil, fmt.Errorf("%w (download it with: cpancompat testers -perl %s)", err, perl)
	}
	defer f.Close()
	byRelease := map[string]*perlReports{}
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1<<20), 1<<20)
	for sc.Scan() {
		var r testerReport
		if err := json.Unmarshal(sc.Bytes(), &r); err != nil {
			return nil, err
		}
		name := r.Dist + "-" + r.Version
		c := byRelease[name]
		if c == nil {
			c = &perlReports{Perl: perl}
			byRelease[name] = c
		}
		linux := r.OSName == "linux"
		switch r.Grade {
		case "pass":
			c.Pass++
			if linux {
				c.LinuxPass++
			}
		case "fail":
			c.Fail++
			if linux {
				c.LinuxFail++
			}
		case "na":
			c.NA++
		case "unknown":
			c.Unknown++
		default:
			return nil, fmt.Errorf("%s: report %s has grade %q", testersFile(perl), r.GUID, r.Grade)
		}
	}
	return byRelease, sc.Err()
}
