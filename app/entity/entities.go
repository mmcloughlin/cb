package entity

import (
	"crypto/sha256"
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/mmcloughlin/goperf/app/change"
	"github.com/mmcloughlin/goperf/app/id"
	"github.com/mmcloughlin/goperf/app/trace"
)

type Person struct {
	Name  string
	Email string
}

type Commit struct {
	SHA        string
	Tree       string
	Parents    []string
	Author     Person
	AuthorTime time.Time
	Committer  Person
	CommitTime time.Time
	Message    string
}

type CommitRef struct {
	SHA string
	Ref string
}

type CommitPosition struct {
	SHA        string
	CommitTime time.Time
	Index      int
}

type CommitIndexRange struct {
	Min int
	Max int
}

func SingleCommitIndexRange(idx int) CommitIndexRange {
	return CommitIndexRange{Min: idx, Max: idx}
}

func (r CommitIndexRange) String() string { return fmt.Sprintf("[%d,%d]", r.Min, r.Max) }

type DataFile struct {
	Name   string
	SHA256 [sha256.Size]byte
}

var datafilenamespace = uuid.MustParse("3e094884-6ffd-4d70-a83f-bc2d241b7344")

func (f *DataFile) UUID() uuid.UUID {
	return id.UUID(datafilenamespace, f.SHA256[:])
}

type Result struct {
	File        *DataFile
	Line        int
	Benchmark   *Benchmark
	Commit      *Commit
	Environment Properties
	Metadata    Properties
	Iterations  uint64
	Value       float64
}

var resultnamespace = uuid.MustParse("0063a4c4-2bdc-4c3b-878b-5c90356013a3")

func (r *Result) UUID() uuid.UUID {
	return id.Strings(resultnamespace, []string{
		r.File.UUID().String(),
		strconv.Itoa(r.Line),
		r.Benchmark.UUID().String(),
	})
}

type Properties map[string]string

var propertiesnamespace = uuid.MustParse("d0c136af-cf22-4f7a-87b3-4a73bfb57489")

func (p Properties) UUID() uuid.UUID {
	return id.KeyValues(propertiesnamespace, p)
}

type Module struct {
	Path    string
	Version string
}

var modulenamespace = uuid.MustParse("24662676-cba4-4241-ab2d-e81de0d407b4")

func (m *Module) UUID() uuid.UUID {
	return id.Strings(modulenamespace, []string{m.Path, m.Version})
}

func (m *Module) String() string {
	if m.Version == "" {
		return m.Path
	}
	return m.Path + "@" + m.Version
}

type Package struct {
	Module       *Module
	RelativePath string
}

var packagenamespace = uuid.MustParse("91e2ea8d-5830-4b70-b26c-68ad426636eb")

func (p *Package) UUID() uuid.UUID {
	return id.Strings(packagenamespace, []string{
		p.Module.UUID().String(),
		p.RelativePath,
	})
}

func (p *Package) ImportPath() string {
	return path.Join(p.Module.Path, p.RelativePath)
}

type Benchmark struct {
	Package    *Package
	FullName   string
	Name       string
	Parameters map[string]string
	Unit       string
}

var benchmarknamespace = uuid.MustParse("51d0f236-a868-48f4-8ef1-8d303e4953e1")

func (b *Benchmark) UUID() uuid.UUID {
	paramsid := id.KeyValues(uuid.Nil, b.Parameters)
	return id.Strings(benchmarknamespace, []string{
		b.Package.UUID().String(),
		b.Name,
		paramsid.String(),
		b.Unit,
	})
}

// Point in a benchmark result timeseries.
type Point struct {
	ResultUUID      uuid.UUID
	BenchmarkUUID   uuid.UUID
	EnvironmentUUID uuid.UUID
	CommitSHA       string
	CommitIndex     int
	Value           float64
}

// Points is a series of measurements.
type Points []*Point

// Values returns the series of point values.
func (p Points) Values() []float64 {
	xs := make([]float64, len(p))
	for i := range p {
		xs[i] = p[i].Value
	}
	return xs
}

type Change struct {
	trace.ID
	change.Change
}
