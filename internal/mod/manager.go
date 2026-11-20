package mod

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync"
	"time"
)

// The services module names are resolved with, as cpanm and cpm use them.
const (
	defaultMetaDB   = "https://cpanmetadb.plackperl.org/v1.0/"
	defaultMetaCPAN = "https://fastapi.metacpan.org/v1/download_url/"
)

// parallelism bounds concurrent downloads and lookups.
const parallelism = 8

// Manager resolves, downloads, and unpacks distributions into a module
// directory.
type Manager struct {
	Dir        string // the module directory
	Mirror     string // CPAN mirror, with a trailing slash
	MirrorOnly bool   // resolve with the mirror's own index only
	MetaDB     string
	MetaCPAN   string
	HTTP       *http.Client
	Log        io.Writer // progress lines

	helper *Helper

	mu        sync.Mutex
	checksums map[string]*checksumSet // by author directory
	index     map[string]indexEntry   // the mirror's 02packages, once loaded
	bundled   map[string]bundledModule
}

type bundledModule struct {
	version string
	ok      bool
}

// NewManager builds a manager from the environment (GPERLMODCACHE,
// GPERLMIRROR, GPERLMIRRORONLY).
func NewManager() (*Manager, error) {
	dir, err := Dir()
	if err != nil {
		return nil, err
	}
	return &Manager{
		Dir:        dir,
		Mirror:     Mirror(),
		MirrorOnly: MirrorOnly(),
		MetaDB:     defaultMetaDB,
		MetaCPAN:   defaultMetaCPAN,
		HTTP:       &http.Client{Timeout: 5 * time.Minute},
		Log:        os.Stderr,
		helper:     &Helper{},
		checksums:  map[string]*checksumSet{},
		bundled:    map[string]bundledModule{},
	}, nil
}

// Close releases the manager's interpreter.
func (m *Manager) Close() error { return m.helper.Close() }

func (m *Manager) logf(format string, args ...any) {
	fmt.Fprintf(m.Log, "gperl: "+format+"\n", args...)
}

// errNotFound reports that a service has no answer for a request.
var errNotFound = errors.New("not found")

// get fetches url, retrying once on a transport error or a server error.
// A 404 is errNotFound.
func (m *Manager) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		body, err := m.getOnce(ctx, url)
		if err == nil || errors.Is(err, errNotFound) || ctx.Err() != nil {
			return body, err
		}
		lastErr = err
	}
	return nil, lastErr
}

func (m *Manager) getOnce(ctx context.Context, url string) (body []byte, err error) {
	resp, err := m.open(ctx, url)
	if err != nil {
		return nil, err
	}
	defer func() { err = errors.Join(err, resp.Close()) }()
	return io.ReadAll(resp)
}

// open starts a GET and returns the body of a 200 response.
func (m *Manager) open(ctx context.Context, url string) (io.ReadCloser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "gperl")
	resp, err := m.HTTP.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode == http.StatusOK {
		return resp.Body, nil
	}
	err = fmt.Errorf("GET %s: %s", url, resp.Status)
	if resp.StatusCode == http.StatusNotFound {
		err = fmt.Errorf("GET %s: %w", url, errNotFound)
	}
	return nil, errors.Join(err, resp.Body.Close())
}

// each runs fn over items with bounded concurrency and returns the first
// error; the remaining items are skipped once one fails.
func each[T any](ctx context.Context, items []T, fn func(context.Context, T) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var (
		wg    sync.WaitGroup
		once  sync.Once
		first error
		slots = make(chan struct{}, parallelism)
	)
	for _, item := range items {
		if ctx.Err() != nil {
			break
		}
		slots <- struct{}{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer func() { <-slots }()
			if ctx.Err() != nil {
				return
			}
			if err := fn(ctx, item); err != nil {
				once.Do(func() { first = err; cancel() })
			}
		}()
	}
	wg.Wait()
	return first
}
