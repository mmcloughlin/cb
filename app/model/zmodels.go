// Code generated by make_models.go. DO NOT EDIT.

package model

import "time"

type Commit struct {
	SHA            string    `firestore:"sha" json:"sha"`
	Tree           string    `firestore:"tree" json:"tree"`
	Parents        []string  `firestore:"parents" json:"parents"`
	AuthorName     string    `firestore:"author_name" json:"author_name"`
	AuthorEmail    string    `firestore:"author_email" json:"author_email"`
	AuthorTime     time.Time `firestore:"author_time" json:"author_time"`
	CommitterName  string    `firestore:"committer_name" json:"committer_name"`
	CommitterEmail string    `firestore:"committer_email" json:"committer_email"`
	CommitTime     time.Time `firestore:"commit_time" json:"commit_time"`
	Message        string    `firestore:"message" json:"message"`
}

func (c *Commit) Type() string { return "commits" }
func (c *Commit) ID() string   { return c.SHA }

type Module struct {
	UUID    string `firestore:"uuid" json:"uuid"`
	Path    string `firestore:"path" json:"path"`
	Version string `firestore:"version" json:"version"`
}

func (m *Module) Type() string { return "modules" }
func (m *Module) ID() string   { return m.UUID }

type Package struct {
	UUID         string `firestore:"uuid" json:"uuid"`
	ModuleUUID   string `firestore:"module_uuid" json:"module_uuid"`
	RelativePath string `firestore:"relative_path" json:"relative_path"`
}

func (p *Package) Type() string { return "packages" }
func (p *Package) ID() string   { return p.UUID }

type Benchmark struct {
	UUID        string            `firestore:"uuid" json:"uuid"`
	PackageUUID string            `firestore:"package_uuid" json:"package_uuid"`
	Name        string            `firestore:"name" json:"name"`
	Unit        string            `firestore:"unit" json:"unit"`
	Parameters  map[string]string `firestore:"parameters" json:"parameters"`
}

func (b *Benchmark) Type() string { return "benchmarks" }
func (b *Benchmark) ID() string   { return b.UUID }
