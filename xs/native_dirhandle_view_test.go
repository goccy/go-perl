//go:build darwin || linux

package xs_test

import (
	"testing"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// The directory stream of a dirhandle (IoDIRP) is native code's to read
// with readdir and its fellows, as IO::Dirent reads names, inodes and types,
// File::StatX takes dirfd for statx and POSIX::2008 perl's my_dirfd for
// openat (HAS_DIRFD); a DIR native code opens and stores
// there (POSIX::2008's fdopendir) is the dirhandle's from then on, and one
// it closes and sets to NULL leaves the dirhandle closed. Before, IoDIRP did
// not compile. The expected string is what perl prints with the same
// module code. Native code's descriptors are the program's, as they are
// for gperl, a program that is the host process.
const dirhandleViewSource = `#include "EXTERN.h"
#include "perl.h"
#include "XSUB.h"
#include <dirent.h>
#include <sys/stat.h>

static DIR *MyDirv_dir(pTHX_ SV *sv) {
    IO *const io = sv_2io(sv);
    DIR *const d = IoDIRP(io);
    if (!d) croak("no dirhandle");
    return d;
}
/* the names of the entries, and with a kind each (d for a directory, f for
 * a file) */
XS(XS_MyDirv_names) {
    dXSARGS;
    DIR *d;
    Direntry_t *e;
    int kinds;
    if (items != 2) croak_xs_usage(cv, "dh, kinds");
    d = MyDirv_dir(aTHX_ ST(0));
    kinds = SvTRUE(ST(1));
    SP -= items;
    while ((e = (Direntry_t *)readdir(d))) {
        if (e->d_name[0] == '.') continue;
        if (kinds)
            mXPUSHs(newSVpvf("%s:%s", e->d_name, e->d_type == DT_DIR ? "d" : e->d_type == DT_REG ? "f" : "?"));
        else
            mXPUSHs(newSVpv(e->d_name, 0));
    }
    PUTBACK;
}
/* the descriptor of the stream is one of the host's, of a directory, by
 * dirfd or by perl's my_dirfd (as POSIX::2008 finds it) */
XS(XS_MyDirv_fd_kind) {
    dXSARGS;
    struct stat st;
    int fd;
    if (items != 2) croak_xs_usage(cv, "dh, perls");
#ifdef HAS_DIRFD
    if (SvTRUE(ST(1)))
        fd = my_dirfd(MyDirv_dir(aTHX_ ST(0)));
    else
#endif
        fd = dirfd(MyDirv_dir(aTHX_ ST(0)));
    ST(0) = sv_2mortal(newSVpv(fd < 0 ? "no fd" : fstat(fd, &st) ? "no stat" : S_ISDIR(st.st_mode) ? "dir" : "other", 0));
    XSRETURN(1);
}
/* telldir, readdir, seekdir back and readdir again read the same entry */
XS(XS_MyDirv_seek_back) {
    dXSARGS;
    DIR *d;
    long pos;
    char first[256];
    Direntry_t *e;
    if (items != 1) croak_xs_usage(cv, "dh");
    d = MyDirv_dir(aTHX_ ST(0));
    pos = telldir(d);
    e = (Direntry_t *)readdir(d);
    if (!e) croak("empty");
    my_strlcpy(first, e->d_name, sizeof first);
    seekdir(d, pos);
    e = (Direntry_t *)readdir(d);
    ST(0) = sv_2mortal(newSVpv(e && strEQ(e->d_name, first) ? "same" : "other", 0));
    XSRETURN(1);
}
/* a directory of native code's opening becomes the dirhandle's */
XS(XS_MyDirv_adopt) {
    dXSARGS;
    DIR *d;
    if (items != 2) croak_xs_usage(cv, "dh, path");
    d = opendir(SvPV_nolen(ST(1)));
    if (!d) croak("opendir: %s", Strerror(errno));
    IoDIRP(sv_2io(ST(0))) = d;
    XSRETURN_EMPTY;
}
/* closed by native code, as POSIX::2008's closedir of a dirhandle */
XS(XS_MyDirv_close) {
    dXSARGS;
    IO *io;
    if (items != 1) croak_xs_usage(cv, "dh");
    io = sv_2io(ST(0));
    PerlDir_close(IoDIRP(io));
    IoDIRP(io) = 0;
    XSRETURN_EMPTY;
}

XS_EXTERNAL(boot_MyDirv) {
    dXSBOOTARGSXSAPIVERCHK;
    newXS_deffile("MyDirv::names", XS_MyDirv_names);
    newXS_deffile("MyDirv::fd_kind", XS_MyDirv_fd_kind);
    newXS_deffile("MyDirv::seek_back", XS_MyDirv_seek_back);
    newXS_deffile("MyDirv::adopt", XS_MyDirv_adopt);
    newXS_deffile("MyDirv::close", XS_MyDirv_close);
    Perl_xs_boot_epilog(aTHX_ ax);
}
`

func TestNativeDirectoryStreamOfADirhandle(t *testing.T) {
	so := buildModuleSource(t, "MyDirv", dirhandleViewSource)
	// The program removes what it made without File::Temp's cleanup, which
	// changes directory: the host process, which follows the program's
	// directory, would be left elsewhere for the tests after this one.
	got := requireModuleWith(t, perl.Config{FS: goperlfs.NewHostFS(), HostProcess: true}, "MyDirv", so, `use File::Temp qw(tempdir);
my $d = tempdir();
for my $f ("a", "b") { open my $fh, ">", "$d/$f" or die; close $fh }
mkdir "$d/c" or die;
my $e = tempdir();
open my $fh, ">", "$e/x" or die; close $fh;
opendir(my $dh, $d) or die;
my @r = join ",", sort(MyDirv::names($dh, 0));
rewinddir $dh;
push @r, join ",", sort(MyDirv::names($dh, 1));
rewinddir $dh;
push @r, MyDirv::fd_kind($dh, 0) . "," . MyDirv::fd_kind($dh, 1), MyDirv::seek_back($dh);
opendir(my $other, $e) or die;
MyDirv::adopt($other, $d);
push @r, join ",", sort grep { !/^\./ } readdir $other;
MyDirv::close($dh);
{ no warnings; push @r, defined(readdir $dh) ? "open" : "closed" }
closedir $other;
unlink "$d/a", "$d/b", "$e/x";
rmdir $_ for "$d/c", $d, $e;
join "|", @r, (-e $d || -e $e ? "left" : "removed")`)
	want := "a,b,c|a:f,b:f,c:d|dir,dir|same|a,b,c|closed|removed"
	if got != want {
		t.Errorf("got  %q\nwant %q", got, want)
	}
}
