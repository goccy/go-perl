package gperl

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"

	perl "github.com/goccy/go-perl"
)

// The running executable as the perl command.
//
// A perl knows the path it can be run by: $^X, and %Config's perlpath and
// startperl, which is what an installation writes on the #! line of the
// programs it installs. The executable that runs the interpreter here is
// not a perl command by its own command line (gperl wants `run` first), so
// it has a second name that is one: a link called perl, in a directory of
// its own beside the standard library (in the user's cache, or where the
// standard library goes when there is none), that points at the
// executable. Started by that name the executable takes its whole command
// line as perl's (see reexec.go).
//
// perl's own programs (prove, perldoc, pod2usage, ...) are beside it, as
// they are beside an installed perl command: %Config names the directory
// of the perl command as theirs (bin, scriptdir), and they name the link on
// their #! line, so that running one runs it on this executable. They come
// from the standard library, which ships them (bin/ in it).
//
// The place of the directory follows from the place of the executable and
// the version of the standard library alone, so a directory that was
// removed is the same directory again after the next run, and a standard
// library of another version (the executable replaced by a newer one) has
// its programs in a directory of its own.

// selfPerlPath returns where the link to this executable is (whether or not
// it exists), the executable it points at and the standard library whose
// programs are beside it; ok is false when the process cannot name its
// executable or there is no standard library.
func selfPerlPath() (link, exe, stdlib string, ok bool) {
	exe, err := os.Executable()
	if err != nil {
		return "", "", "", false
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return "", "", "", false
	}
	if stdlib, err = perl.ExtractStdlib(); err != nil {
		return "", "", "", false
	}
	// The standard library is named by its directory's own name, which
	// says what it holds (stdlib-<digest>): the directory is reached by
	// different paths (another home directory whose caches are the same
	// ones), and a process extracts it once, so a path would name another
	// link in a process whose environment has changed since.
	sum := sha256.Sum256([]byte(exe + "\x00" + filepath.Base(stdlib)))
	return filepath.Join(filepath.Dir(stdlib), "perl-"+hex.EncodeToString(sum[:8]), "perl"), exe, stdlib, true
}

// selfPerl returns the path that runs this executable as the perl command,
// making the link and the programs beside it when the link is not there;
// "" when there is no such path on this system.
func selfPerl() string {
	link, exe, stdlib, ok := selfPerlPath()
	if !ok {
		return ""
	}
	if target, err := os.Readlink(link); err == nil && target == exe {
		return link
	}
	dir := filepath.Dir(link)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ""
	}
	// Whatever is at a name is not what belongs there: a new file is made
	// beside it and renamed over it, so that the name has a whole file at
	// every moment, also while several processes do this at once. The link
	// comes last: where it is, the programs are.
	tmp, err := os.MkdirTemp(dir, "new-")
	if err != nil {
		return ""
	}
	defer os.RemoveAll(tmp)
	if err := placePrograms(filepath.Join(stdlib, "bin"), tmp, dir, link); err != nil {
		return ""
	}
	fresh := filepath.Join(tmp, "perl")
	if err := os.Symlink(exe, fresh); err != nil {
		return ""
	}
	if err := os.Rename(fresh, link); err != nil {
		return ""
	}
	return link
}

// placePrograms puts each program of the directory from into dir, by way of
// the scratch directory tmp, with a #! line naming perlPath in place of the
// one it has.
func placePrograms(from, tmp, dir, perlPath string) error {
	entries, err := os.ReadDir(from)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.Type().IsRegular() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(from, e.Name()))
		if err != nil {
			return err
		}
		fresh := filepath.Join(tmp, e.Name())
		if err := os.WriteFile(fresh, withInterpreter(data, perlPath), 0o755); err != nil {
			return err
		}
		if err := os.Rename(fresh, filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

// withInterpreter returns the text of a program with the interpreter line
// (#!, its first line, as the system reads it) naming path.
func withInterpreter(program []byte, path string) []byte {
	line := []byte("#!" + path + "\n")
	if !bytes.HasPrefix(program, []byte("#!")) {
		return append(line, program...)
	}
	if i := bytes.IndexByte(program, '\n'); i >= 0 {
		return append(line, program[i+1:]...)
	}
	return line
}

// startedAsPerl reports whether the process was started by the name
// selfPerl gives the executable: the name the system passed as the
// command's own (the path on a #! line, or the path a program ran). That
// is also the link in the cache of another home directory: a program's #!
// line names the link of the home directory it was written under, and a
// program run with HOME changed (as a test runs one) has another cache. Its
// directory has the same name, which follows from the executable and the
// standard library alone, and it is a link to this executable.
func startedAsPerl() bool {
	if len(os.Args) == 0 || !filepath.IsAbs(os.Args[0]) || filepath.Base(os.Args[0]) != "perl" {
		return false
	}
	link, exe, _, ok := selfPerlPath()
	if !ok {
		return false
	}
	arg := filepath.Clean(os.Args[0])
	if arg == link {
		return true
	}
	if filepath.Base(filepath.Dir(arg)) != filepath.Base(filepath.Dir(link)) {
		return false
	}
	target, err := filepath.EvalSymlinks(arg)
	return err == nil && target == exe
}
