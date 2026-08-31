package perl_test

import (
	"context"
	"fmt"
	"sync"
	"testing"

	perl "github.com/goccy/go-perl"
)

// TestCloneInheritsAndDiverges pins the Clone contract: a clone sees
// everything the prototype compiled and defined, and from that point the
// two diverge privately.
func TestCloneInheritsAndDiverges(t *testing.T) {
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	ctx := context.Background()

	if r, err := p.Eval(ctx, `our $state = 'proto'; sub whoami { $main::state } 1;`); err != nil || r.Error != nil {
		t.Fatalf("prepare prototype: err=%v error=%v", err, r.Error)
	}
	if err := p.Bind("go_add", func(args []perl.Value) ([]perl.Value, error) {
		a := scalarOf(t, args[0])
		b := scalarOf(t, args[1])
		return []perl.Value{perl.ValueOf(a.Int() + b.Int())}, nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}

	c, err := p.Clone()
	if err != nil {
		t.Fatalf("Clone: %v", err)
	}
	defer c.Close()

	// The clone inherits compiled subs, package state, and Go bindings.
	if r, err := c.Eval(ctx, `whoami()`); err != nil || r.Error != nil || resultStr(r) != "proto" {
		t.Fatalf("clone inherit: err=%v ok=%v result=%q error=%v", err, (r.Error == nil), resultStr(r), r.Error)
	}
	if r, err := c.Eval(ctx, `go_add(20, 22)`); err != nil || r.Error != nil || resultStr(r) != "42" {
		t.Fatalf("clone binding: err=%v result=%q error=%v", err, resultStr(r), r.Error)
	}

	// Divergence: writes in one are invisible in the other.
	if r, err := c.Eval(ctx, `$main::state = 'clone'; whoami()`); err != nil || r.Error != nil || resultStr(r) != "clone" {
		t.Fatalf("clone write: err=%v result=%q error=%v", err, resultStr(r), r.Error)
	}
	if r, err := p.Eval(ctx, `whoami()`); err != nil || r.Error != nil || resultStr(r) != "proto" {
		t.Fatalf("prototype isolated: err=%v result=%q error=%v", err, resultStr(r), r.Error)
	}

	// A second clone starts from the image, not from the first clone.
	c2, err := p.Clone()
	if err != nil {
		t.Fatalf("second Clone: %v", err)
	}
	defer c2.Close()
	if r, err := c2.Eval(ctx, `whoami()`); err != nil || r.Error != nil || resultStr(r) != "proto" {
		t.Fatalf("second clone: err=%v result=%q error=%v", err, resultStr(r), r.Error)
	}
}

// TestClonesRunConcurrently exercises clones on parallel goroutines.
func TestClonesRunConcurrently(t *testing.T) {
	p, err := perl.New(perl.Config{})
	if err != nil {
		t.Fatal(err)
	}
	defer p.Close()
	ctx := context.Background()
	if r, err := p.Eval(ctx, `sub work { my $n = shift; my $s = 0; $s += $_ for 1 .. $n; $s } 1;`); err != nil || r.Error != nil {
		t.Fatalf("prepare: err=%v error=%v", err, r.Error)
	}

	const workers = 4
	clones := make([]*perl.Perl, workers)
	for i := range clones {
		c, err := p.Clone()
		if err != nil {
			t.Fatalf("Clone %d: %v", i, err)
		}
		defer c.Close()
		clones[i] = c
	}
	var wg sync.WaitGroup
	errs := make(chan error, workers)
	for _, c := range clones {
		wg.Add(1)
		go func(c *perl.Perl) {
			defer wg.Done()
			for i := 0; i < 25; i++ {
				r, err := c.Eval(ctx, `work(1000)`)
				if err != nil || r.Error != nil || resultStr(r) != "500500" {
					errs <- fmt.Errorf("work: err=%v ok=%v result=%q error=%v", err, (r.Error == nil), resultStr(r), r.Error)
					return
				}
			}
		}(c)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatal(err)
	}
}
