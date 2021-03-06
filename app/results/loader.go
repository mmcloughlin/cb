// Package results implements loading of benchmark results from data files.
package results

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mmcloughlin/goperf/app/entity"
	"github.com/mmcloughlin/goperf/app/repo"
	"github.com/mmcloughlin/goperf/internal/errutil"
	"github.com/mmcloughlin/goperf/pkg/cfg"
	"github.com/mmcloughlin/goperf/pkg/fs"
	"github.com/mmcloughlin/goperf/pkg/mod"
	"github.com/mmcloughlin/goperf/pkg/parse"
)

// Keys defines which configuration keys hold critical parameters in results
// loading.
type Keys struct {
	ToolchainRef string // git ref for the go toolchain version under test
	Module       string // go module under test
	Package      string // package under test
}

// All returns all special keys.
func (k Keys) All() []string {
	return []string{
		k.ToolchainRef,
		k.Module,
		k.Package,
	}
}

// DefaultKeys defines the default configuration keys holding critical result
// parameters.
var DefaultKeys = Keys{
	ToolchainRef: "toolchain-ref",
	Module:       "suite-mod",
	Package:      "pkg",
}

// Loader loads benchmark result files and associated data.
type Loader struct {
	fs      fs.Readable
	rev     repo.Revisions
	mod     mod.Infoer
	keys    Keys
	envtags []cfg.Tag
}

// LoaderOption configures a Loader.
type LoaderOption func(*Loader)

// WithFilesystem configures a Loader to read data files from the given filesystem.
func WithFilesystem(r fs.Readable) LoaderOption {
	return func(l *Loader) { l.fs = r }
}

// WithRevisions configures a Loader to read commit data from the supplied
// Revisions fetcher.
func WithRevisions(r repo.Revisions) LoaderOption {
	return func(l *Loader) { l.rev = r }
}

// WithModuleInfo configures how a Loader fetches module information.
func WithModuleInfo(i mod.Infoer) LoaderOption {
	return func(l *Loader) { l.mod = i }
}

// WithKeys configures which benchmark configuration keys are special.
func WithKeys(k Keys) LoaderOption {
	return func(l *Loader) { l.keys = k }
}

// WithEnvironmentTags configures which configuration tags are are considered
// environment, meaning that their value has an affect on the benchmark result.
func WithEnvironmentTags(tags ...cfg.Tag) LoaderOption {
	return func(l *Loader) { l.envtags = tags }
}

// NewLoader builds a new benchmark loader.
func NewLoader(opts ...LoaderOption) (*Loader, error) {
	l := &Loader{
		rev:     repo.NewRevisionsCache(repo.Go(http.DefaultClient), 16),
		mod:     mod.NewInfoCache(mod.NewOfficialModuleProxy(http.DefaultClient), 16),
		keys:    DefaultKeys,
		envtags: []cfg.Tag{cfg.TagPerfCritical},
	}
	for _, opt := range opts {
		opt(l)
	}
	if l.fs == nil {
		return nil, errors.New("must configure a filesystem")
	}
	return l, nil
}

// Load the named benchmark file.
func (l *Loader) Load(ctx context.Context, name string) ([]*entity.Result, error) {
	return l.load(ctx, name)
}

func (l *Loader) load(ctx context.Context, name string) (_ []*entity.Result, err error) {
	// Open the input data file.
	f, err := l.fs.Open(ctx, name)
	if err != nil {
		return nil, err
	}
	defer errutil.CheckClose(&err, f)

	// Hash the file while reading it.
	h := sha256.New()
	r := io.TeeReader(f, h)

	// Parse.
	collection, err := parse.Reader(r)
	if err != nil {
		return nil, err
	}

	// Construct DataFile object.
	datafile := &entity.DataFile{Name: name}
	h.Sum(datafile.SHA256[:0])

	// Process results.
	output := make([]*entity.Result, 0, len(collection.Results))
	for _, result := range collection.Results {
		out, err := l.convert(ctx, result)
		if err != nil {
			return nil, err
		}
		out.File = datafile
		output = append(output, out)
	}

	return output, nil
}

// conert the parsed result into a model Result.
func (l *Loader) convert(ctx context.Context, r *parse.Result) (*entity.Result, error) {
	// Lookup commit.
	commit, err := l.commit(ctx, r)
	if err != nil {
		return nil, err
	}

	// Build benchmark.
	bench, err := l.benchmark(ctx, r)
	if err != nil {
		return nil, err
	}

	// Separate environment labels from metadata.
	env := l.environment(r)

	meta := map[string]string{}
	for k, v := range r.Labels {
		if _, ok := env[k]; !ok {
			meta[k] = v
		}
	}

	l.deletekeys(env)
	l.deletekeys(meta)

	return &entity.Result{
		Line:        r.Line,
		Benchmark:   bench,
		Commit:      commit,
		Environment: env,
		Metadata:    meta,
		Iterations:  r.Iterations,
		Value:       r.Value,
	}, nil
}

// commit looks up the commit associated with the given result.
func (l *Loader) commit(ctx context.Context, r *parse.Result) (*entity.Commit, error) {
	ref, err := lookup(r.Labels, l.keys.ToolchainRef)
	if err != nil {
		return nil, err
	}
	return l.rev.Revision(ctx, ref)
}

// benchmark builds the benchmark object corresponding to this result.
func (l *Loader) benchmark(ctx context.Context, r *parse.Result) (*entity.Benchmark, error) {
	// Load module path, version and package.
	module, err := lookup(r.Labels, l.keys.Module)
	if err != nil {
		return nil, err
	}

	m := mod.Parse(module)

	pkgpath, err := lookup(r.Labels, l.keys.Package)
	if err != nil {
		return nil, err
	}

	// Resolve canonical version.
	v, err := l.version(ctx, m.Path, m.Version)
	if err != nil {
		return nil, err
	}

	// Extract package path relative to module.
	relpath := pkgpath
	if !mod.IsMetaPackage(m.Path) {
		relpath, err = rel(m.Path, pkgpath)
		if err != nil {
			return nil, err
		}
	}

	// Build entity objects.
	mod := &entity.Module{
		Path:    m.Path,
		Version: v,
	}

	pkg := &entity.Package{
		Module:       mod,
		RelativePath: relpath,
	}

	bench := &entity.Benchmark{
		Package:    pkg,
		FullName:   r.FullName,
		Name:       r.Name,
		Parameters: r.Parameters,
		Unit:       r.Unit,
	}

	return bench, nil
}

// version fetches the canonical module version.
func (l *Loader) version(ctx context.Context, path, rev string) (string, error) {
	// Special case for meta packages like standard library.
	if mod.IsMetaPackage(path) {
		if rev != "" {
			return "", fmt.Errorf("unexpected version for meta package %q: got %q", path, rev)
		}
		return "", nil
	}

	// Fetch version.
	info, err := l.mod.Info(ctx, path, rev)
	if err != nil {
		return "", err
	}

	return info.Version, nil
}

// environment extracts the environment fields from the benchmark labels.
func (l *Loader) environment(r *parse.Result) map[string]string {
	groups := groupbytags(r.Labels)
	env := map[string]string{}
	for _, tag := range l.envtags {
		for k, v := range groups[tag] {
			env[k] = v
		}
	}
	return env
}

// deletekeys clears special keys from the supplied properties.
func (l *Loader) deletekeys(p entity.Properties) {
	for _, key := range l.keys.All() {
		delete(p, key)
	}
}

// lookup key k in map m, returning a human-readable error.
func lookup(m map[string]string, k string) (string, error) {
	v, ok := m[k]
	if !ok {
		return "", fmt.Errorf("key %q missing", k)
	}
	return v, nil
}

// rel returns the path to pkg relative to mod.
func rel(mod, pkg string) (string, error) {
	if !strings.HasPrefix(pkg, mod) {
		return "", fmt.Errorf("package %q does not belong to module %q", pkg, mod)
	}
	if pkg == mod {
		return "", nil
	}
	rest := pkg[len(mod):]
	if rest[0] != '/' {
		return "", errors.New("expect path separator")
	}
	return rest[1:], nil
}

// groupbytags parses tags from benchmark labels and returns a mapping from tags
// to associated labels.
func groupbytags(labels map[string]string) map[cfg.Tag]map[string]string {
	groups := map[cfg.Tag]map[string]string{}
	for key, value := range labels {
		v, tags := cfg.ParseValueTags(value)
		for _, tag := range tags {
			if _, ok := groups[tag]; !ok {
				groups[tag] = map[string]string{}
			}
			groups[tag][key] = v
		}
	}
	return groups
}
