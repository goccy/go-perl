package perl_test

import (
	"os"
	"path/filepath"
	"testing"

	perl "github.com/goccy/go-perl"
	"github.com/goccy/go-perl/internal"
)

// The name of the architecture is the host's, marked as go-perl's: %Config
// says so (archname and myarchname), and the architecture subdirectory perl
// adds to @INC with a PERL5LIB directory has that name.
func TestTheArchitectureIsTheHosts(t *testing.T) {
	want := internal.ArchName()
	five := t.TempDir()
	if err := os.Mkdir(filepath.Join(five, want), 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := hostConfig(t)
	cfg.Env = []string{"PERL5LIB=" + five}
	r := runProgram(t, cfg, perl.Program{
		Code: code(
			`use Config;`,
			`print "$Config{archname} $Config{myarchname}\n";`,
			`print join(",", grep { index($_, $ARGV[0]) == 0 } @INC), "\n";`,
		),
		Args: []string{five},
	}, "")
	r.want(t, 0, want+" "+want+"\n"+filepath.Join(five, want)+","+five+"\n")
}
