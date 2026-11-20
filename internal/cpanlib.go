package internal

import "io/fs"

// cpanLib is the module tree the go-perl/cpan package carries.
var cpanLib fs.FS

// SetCPANLib installs the module tree every new instance gets on @INC.
// Called by the init of a generated go-perl/cpan package; not for
// application use.
func SetCPANLib(fsys fs.FS) { cpanLib = fsys }

// CPANLib returns the installed module tree, nil when the binary carries
// none.
func CPANLib() fs.FS { return cpanLib }
