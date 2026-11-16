//go:build darwin || linux

package xs_test

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	perl "github.com/goccy/go-perl"
	goperlfs "github.com/goccy/go-perl/fs"
)

// threadStatePerl makes an instance with cfg (FS and StdlibDir filled in)
// that has testdata/threadstate's module registered and loaded.
func threadStatePerl(t testing.TB, cfg perl.Config) *perl.Perl {
	t.Helper()
	cc, err := exec.LookPath("cc")
	if err != nil {
		t.Skip("no host C compiler available")
	}
	so := filepath.Join(t.TempDir(), "ThreadState.so")
	args := []string{"-shared", "-fPIC", "-I", "sdk/include", `-DXS_VERSION="1.00"`, "-o", so,
		"testdata/threadstate/ThreadState.c"}
	if runtime.GOOS == "darwin" {
		switch runtime.GOARCH {
		case "arm64":
			args = append([]string{"-arch", "arm64"}, args...)
		case "amd64":
			args = append([]string{"-arch", "x86_64"}, args...)
		}
	}
	if out, err := exec.Command(cc, args...).CombinedOutput(); err != nil {
		t.Fatalf("compile ThreadState.so: %v\n%s", err, out)
	}
	stdlib, err := perl.ExtractStdlib()
	if err != nil {
		t.Fatalf("extract stdlib: %v", err)
	}
	cfg.FS = goperlfs.NewHostFS()
	cfg.StdlibDir = stdlib
	p, err := perl.New(cfg)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	t.Cleanup(func() { p.Close() })
	loadXSModule(t, p, "ThreadState", so)
	libdir := t.TempDir()
	if err := os.WriteFile(filepath.Join(libdir, "ThreadState.pm"),
		[]byte("package ThreadState;\nuse XSLoader;\nXSLoader::load(__PACKAGE__);\n1;\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	evalStr(t, p, `unshift @INC, '`+libdir+`'; require ThreadState; 1`)
	return p
}

func evalStr(t testing.TB, p *perl.Perl, code string) string {
	t.Helper()
	r, err := p.Eval(context.Background(), code)
	if err != nil {
		t.Fatalf("Eval %q: %v", code, err)
	}
	if r.Error != nil {
		t.Fatalf("Eval %q died: %v", code, r.Error)
	}
	return resultStr(r)
}

// busyScheduler keeps every P of the Go scheduler busy with goroutines that
// yield, until the test ends: a goroutine that waits (a sleep, a collection)
// resumes on whichever thread is free.
func busyScheduler(t *testing.T) {
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < runtime.GOMAXPROCS(0); i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for {
				select {
				case <-stop:
					return
				default:
					runtime.Gosched()
				}
			}
		}()
	}
	t.Cleanup(func() { close(stop); wg.Wait() })
}

// A module's thread-local variable set in one call is there in the next,
// as on perl, where every call of a module runs on the program's thread:
// an instance that runs native code runs on one thread for its life,
// whichever goroutine calls it, while the scheduler is busy and collections
// run between the calls. Before, the calls ran on whatever thread the
// caller's goroutine had, and the value was lost whenever that changed.
func TestNativeThreadLocalStateAcrossCalls(t *testing.T) {
	p := threadStatePerl(t, perl.Config{})
	busyScheduler(t)
	first := evalStr(t, p, `ThreadState::thread_id()`)
	const n = 200
	for i := 1; i <= n; i++ {
		var set string
		if i%3 == 0 {
			// a call from another goroutine
			done := make(chan struct{})
			var r perl.Result
			var err error
			go func() {
				r, err = p.Eval(context.Background(), `ThreadState::tls_set(`+strconv.Itoa(i)+`)`)
				close(done)
			}()
			<-done
			if err != nil || r.Error != nil {
				t.Fatalf("call %d from another goroutine: %v %v", i, err, r.Error)
			}
			set = resultStr(r)
		} else {
			set = evalStr(t, p, `ThreadState::tls_set(`+strconv.Itoa(i)+`)`)
		}
		if i%2 == 0 {
			time.Sleep(time.Millisecond)
		} else {
			runtime.GC()
		}
		got := evalStr(t, p, `ThreadState::thread_id() . ":" . ThreadState::tls_get()`)
		tid, val, _ := strings.Cut(got, ":")
		if set != first || tid != first || val != strconv.Itoa(i) {
			t.Fatalf("call %d: set on thread %s, read on %s (the first call ran on %s): value %s", i, set, tid, first, val)
		}
	}
}

// The instance's thread runs the Go functions the program calls, and an
// entry they make into the instance runs there in place; a context that
// ends stops the program on that thread, which keeps its native state.
func TestNativeThreadReentryAndCancellation(t *testing.T) {
	p := threadStatePerl(t, perl.Config{})
	ctx := context.Background()
	if err := p.Bind("reenter", func(args []perl.Value) ([]perl.Value, error) {
		r, err := p.Eval(ctx, `ThreadState::thread_id()`)
		if err != nil {
			return nil, err
		}
		if r.Error != nil {
			return nil, r.Error
		}
		return []perl.Value{perl.ValueOf(resultStr(r))}, nil
	}); err != nil {
		t.Fatalf("Bind: %v", err)
	}
	if got := evalStr(t, p, `my $a = ThreadState::thread_id(); my ($b) = reenter(); $a eq $b ? "same" : "$a $b"`); got != "same" {
		t.Fatalf("re-entry: %s", got)
	}
	tctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := p.Eval(tctx, `ThreadState::tls_set(77); my $n = 0; while (1) { $n++ }`)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cancelled eval: %v", err)
	}
	if elapsed := time.Since(start); elapsed > 5*time.Second {
		t.Fatalf("cancellation took %v", elapsed)
	}
	if got := evalStr(t, p, `ThreadState::tls_get()`); got != "77" {
		t.Fatalf("after the cancellation: %s", got)
	}
}

// BenchmarkInstanceThreadEval measures an Eval of a constant: on an
// instance with native code, which hands the call to its thread and waits,
// and on one without, which runs it on the caller's goroutine. The
// difference is the hand-off's cost.
func BenchmarkInstanceThreadEval(b *testing.B) {
	ctx := context.Background()
	for _, bc := range []struct {
		name string
		p    func(testing.TB) *perl.Perl
	}{
		{"own_thread", func(tb testing.TB) *perl.Perl { return threadStatePerl(tb, perl.Config{}) }},
		{"caller_goroutine", func(tb testing.TB) *perl.Perl { return newHostPerl(tb) }},
	} {
		b.Run(bc.name, func(b *testing.B) {
			p := bc.p(b)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if _, err := p.Eval(ctx, `1`); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
