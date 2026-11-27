package internal

import "io/fs"

// cpanLib is the module tree the go-perl/cpan package carries, and
// cpanLibSets how many times one was installed.
var (
	cpanLib     fs.FS
	cpanLibSets uint64
)

// SetCPANLib installs the module tree new instances get on @INC (see the
// go-perl package's cpanLibInc).
// Called by the init of a generated go-perl/cpan package; not for
// application use.
func SetCPANLib(fsys fs.FS) { cpanLib, cpanLibSets = fsys, cpanLibSets+1 }

// CPANLib returns the installed module tree, nil when the binary carries
// none, and a number that changes whenever another tree is installed, for
// what is built from the tree to be kept until then.
func CPANLib() (fs.FS, uint64) { return cpanLib, cpanLibSets }
