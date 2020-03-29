package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"path/filepath"
	"sync"
	"time"

	"github.com/mmcloughlin/cb/pkg/fs"
)

func NewStatic(filesys fs.Readable) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		name := r.URL.Path

		info, err := filesys.Stat(ctx, name)
		if err != nil {
			Error(w, err)
			return
		}

		b, err := fs.ReadFile(ctx, filesys, name)
		if err != nil {
			Error(w, err)
			return
		}

		http.ServeContent(w, r, name, info.ModTime, bytes.NewReader(b))
	})
}

func Error(w http.ResponseWriter, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}

type proxysingleurl struct {
	u *url.URL
	c *http.Client

	mu      sync.Mutex
	data    []byte
	modtime time.Time
}

// ProxySingleURL is a handler that serves the content at the given URL.
func ProxySingleURL(u *url.URL) http.Handler {
	return &proxysingleurl{
		u: u,
		c: http.DefaultClient,
	}
}

func (p *proxysingleurl) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != p.u.Path {
		http.NotFound(w, r)
		return
	}

	// Fetch if necessary.
	if !p.fetched() {
		if err := p.fetch(r.Context()); err != nil {
			Error(w, err)
			return
		}
	}

	// Serve.
	http.ServeContent(w, r, filepath.Base(p.u.Path), p.modtime, bytes.NewReader(p.data))
}

func (p *proxysingleurl) fetch(ctx context.Context) error {
	// Make HTTP request.
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, p.u.String(), nil)
	if err != nil {
		return err
	}

	res, err := p.c.Do(request)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	data, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	// Extract last modified date from headers, if possible.
	modtime := time.Now()
	if t, err := http.ParseTime(res.Header.Get("Last-Modified")); err == nil {
		modtime = t
	}

	// Set.
	p.mu.Lock()
	defer p.mu.Unlock()
	p.data = data
	p.modtime = modtime

	return nil
}

func (p *proxysingleurl) fetched() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.data != nil
}
