//go:build darwin || linux

package xs_test

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// A native call that looks at PL_op (dXSTARG, GIMME_V) does work that does
// not grow with the program around the running op: once a tree is
// described, its descriptions last, and the baselines a hand-out gives
// (what a write into a shadow is told by) stop at the ops of finished
// trees. The main program a run starts is finished after INIT, as perl's
// S_run_body finishes it. Before, every such call gave a baseline to every
// op the running op leads to (the whole program: 22068 shadows at 2000
// statements), the first call kept every op of the tree it described among
// them, and the main program of a run was never finished. The
// module counts the shadows with a baseline after it looked at PL_op, an
// SDK internal: this is go-perl's own cost, which perl has no counterpart
// of.
const opHandoutCostSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"

XS(XS_MyCost_shadows) {
    dXSARGS;
    int n = 0;
    goperl_opreg_ent_t *e;
    PERL_UNUSED_VAR(items);
    (void)PL_op->op_private;
    for (e = goperl_opreg_v ? goperl_opreg_v->active : 0; e; e = e->active_next) n++;
    ST(0) = sv_2mortal(newSViv(n));
    XSRETURN(1);
}
XS_EXTERNAL(boot_MyCost) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyCost::shadows", XS_MyCost_shadows);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

// The counts of three calls (the first describes the tree), from a loop in
// the code and from a loop in a sub, with statements statements around.
func opHandoutCostCode(statements int) string {
	never := strings.Repeat("if ($never) { $never++ } ", statements)
	return `our $never; my (@main, @sub);
for my $i (1..3) { push @main, MyCost::shadows() }
` + never + `
sub big { my @r; for my $i (1..3) { push @r, MyCost::shadows() } ` + never + ` @r }
@sub = big();
join "|", join(",", @main), join(",", @sub);`
}

// The counts from the main program of a run (RunProgram, as gperl runs a
// file).
func opHandoutCostRun(t *testing.T, so string, statements int) string {
	t.Helper()
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	var out, errOut bytes.Buffer
	p, err := perl.New(perl.Config{FS: goperlfs.NewHostFS(), StdlibDir: stdlib, NativeModules: true,
		Stdout: &out, Stderr: &errOut})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer p.Close()
	loadXSModule(t, p, "MyCost", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "MyCost.pm"),
		[]byte("package MyCost;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	status, err := p.RunProgram(context.Background(), perl.Program{Inc: []string{libdir}, Code: []string{
		"use MyCost;", opHandoutCostCode(statements), "print join(\"|\", join(\",\", @main), join(\",\", @sub));",
	}})
	if err != nil || status != 0 {
		t.Fatalf("RunProgram: status %d, %v (stderr %q)", status, err, errOut.String())
	}
	return out.String()
}

func TestNativeCallCostDoesNotGrowWithTheProgram(t *testing.T) {
	so := buildModuleSource(t, "MyCost", opHandoutCostSource)
	evalCounts := func(statements int) string {
		return requireModule(t, "MyCost", so, `my $r = eval q{`+opHandoutCostCode(statements)+`}; die $@ if $@; $r`)
	}
	for _, c := range []struct {
		name   string
		counts func(int) string
	}{
		{"string eval", evalCounts},
		{"main program", func(n int) string { return opHandoutCostRun(t, so, n) }},
	} {
		small, big := c.counts(10), c.counts(2000)
		if small != big {
			t.Errorf("%s: shadows given baselines per call: %q with 10 statements, %q with 2000",
				c.name, small, big)
		}
		for _, part := range strings.FieldsFunc(big, func(r rune) bool { return r == '|' || r == ',' }) {
			if len(part) > 2 {
				t.Errorf("%s: a call gave %s shadows baselines (%q)", c.name, part, big)
			}
		}
	}
}
