package internal

import (
	"bytes"
	"errors"
	"slices"
	"strconv"
	"testing"

	"github.com/goccy/perlwasm2go/base"
)

// recordingIDs records the calls a hostChanges makes.
type recordingIDs struct{ calls []string }

func (r *recordingIDs) record(name string, ids ...int) error {
	r.calls = append(r.calls, name)
	for _, id := range ids {
		r.calls = append(r.calls, strconv.Itoa(id))
	}
	return nil
}

func (r *recordingIDs) Setuid(uid int) error          { return r.record("setuid", uid) }
func (r *recordingIDs) Seteuid(euid int) error        { return r.record("seteuid", euid) }
func (r *recordingIDs) Setreuid(ruid, euid int) error { return r.record("setreuid", ruid, euid) }
func (r *recordingIDs) Setresuid(ruid, euid, suid int) error {
	return r.record("setresuid", ruid, euid, suid)
}
func (r *recordingIDs) Setgid(gid int) error          { return r.record("setgid", gid) }
func (r *recordingIDs) Setegid(egid int) error        { return r.record("setegid", egid) }
func (r *recordingIDs) Setregid(rgid, egid int) error { return r.record("setregid", rgid, egid) }
func (r *recordingIDs) Setresgid(rgid, egid, sgid int) error {
	return r.record("setresgid", rgid, egid, sgid)
}
func (r *recordingIDs) Setgroups(gids []int) error { return r.record("setgroups", gids...) }

// TestHostChangesUseTheRegisteredIDs pins that every id change and the
// groups go through the registered way, with the ids the guest gave (-1
// kept as it is).
func TestHostChangesUseTheRegisteredIDs(t *testing.T) {
	ids := &recordingIDs{}
	changes := hostChanges{ProcessChanges: base.HostProcessChanges(), ids: ids}
	for _, change := range []base.IDChange{
		{Call: base.IDSetuid, IDs: [3]int32{1, -1, -1}},
		{Call: base.IDSeteuid, IDs: [3]int32{2, -1, -1}},
		{Call: base.IDSetreuid, IDs: [3]int32{-1, 3, -1}},
		{Call: base.IDSetresuid, IDs: [3]int32{1, 2, 3}},
		{Call: base.IDSetgid, IDs: [3]int32{4, -1, -1}},
		{Call: base.IDSetegid, IDs: [3]int32{5, -1, -1}},
		{Call: base.IDSetregid, IDs: [3]int32{4, -1, -1}},
		{Call: base.IDSetresgid, IDs: [3]int32{4, 5, 6}},
	} {
		if err := changes.SetIDs(change); err != nil {
			t.Fatalf("SetIDs(%+v): %v", change, err)
		}
	}
	if err := changes.SetGroups([]int32{7, 8}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"setuid", "1",
		"seteuid", "2",
		"setreuid", "-1", "3",
		"setresuid", "1", "2", "3",
		"setgid", "4",
		"setegid", "5",
		"setregid", "4", "-1",
		"setresgid", "4", "5", "6",
		"setgroups", "7", "8",
	}
	if !slices.Equal(ids.calls, want) {
		t.Fatalf("calls = %q, want %q", ids.calls, want)
	}
	if err := changes.SetIDs(base.IDChange{Call: base.IDCall(99)}); !errors.Is(err, errors.ErrUnsupported) {
		t.Fatalf("SetIDs of an unknown call: %v, want ErrUnsupported", err)
	}
}

// TestForkFamilyFailuresGoToTheEnd pins where the family's failures that
// no caller takes go: to the end of the family (the first process's
// Close), or, once that is over, to the first process's standard error.
func TestForkFamilyFailuresGoToTheEnd(t *testing.T) {
	var stderr bytes.Buffer
	family := &forkFamily{origin: &Perl{opts: InstanceOptions{Stderr: &stderr}}}
	first := errors.New("first")
	family.fail(failHostSIGCHLD, first)
	family.fail(failHostSIGCHLD, errors.New("second"))
	family.fail(failHostSIGCHLD, errors.New("third"))
	// Of a kind the first is kept and the others are counted; one of
	// another kind is kept besides.
	other := errors.New("other")
	family.fail(failAdoption, other)
	err := family.end()
	if !errors.Is(err, first) || !errors.Is(err, other) {
		t.Fatalf("end: %v, want %v and %v", err, first, other)
	}
	// In the order of the kinds' names.
	want := string(failAdoption) + ": other\n" + string(failHostSIGCHLD) + ": first (and 2 more failures)"
	if err.Error() != want {
		t.Fatalf("end: %q, want %q", err, want)
	}
	family.fail(failAdoption, errors.New("late"))
	if got, want := stderr.String(), "perl: "+string(failAdoption)+": late\n"; got != want {
		t.Fatalf("stderr: %q, want %q", got, want)
	}
}

// TestForkFamilyMember pins that the family's forker finds the instance a
// fork is made by: the first process, or a running pseudo-process.
func TestForkFamilyMember(t *testing.T) {
	origin := &Perl{m: &Module{g: &base.Module{}}}
	family := &forkFamily{origin: origin}
	child := &Perl{m: &Module{g: &base.Module{}}}
	if !family.start(child, 1000) {
		t.Fatal("start refused")
	}
	if got := family.member(origin.m.g); got != origin {
		t.Fatalf("member(origin) = %p, want %p", got, origin)
	}
	if got := family.member(child.m.g); got != child {
		t.Fatalf("member(child) = %p, want %p", got, child)
	}
	family.ended(child)
	if got := family.member(child.m.g); got != nil {
		t.Fatalf("member(ended child) = %p, want nil", got)
	}
}
