package db_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/mmcloughlin/goperf/app/db/dbtest"
	"github.com/mmcloughlin/goperf/app/entity"
	"github.com/mmcloughlin/goperf/app/internal/fixture"
)

func TestDBCommit(t *testing.T) {
	db := dbtest.Open(t)

	// Store.
	ctx := context.Background()
	err := db.StoreCommit(ctx, fixture.Commit)
	if err != nil {
		t.Fatal(err)
	}

	// Find.
	got, err := db.FindCommitBySHA(ctx, fixture.Commit.SHA)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(fixture.Commit, got); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}

func TestDBCommitBatch(t *testing.T) {
	db := dbtest.Open(t)

	// Store in batch mode.
	ctx := context.Background()
	err := db.StoreCommits(ctx, []*entity.Commit{fixture.Commit})
	if err != nil {
		t.Fatal(err)
	}

	// Find.
	got, err := db.FindCommitBySHA(ctx, fixture.Commit.SHA)
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(fixture.Commit, got); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}

func TestDBModule(t *testing.T) {
	db := dbtest.Open(t)

	// Store.
	ctx := context.Background()
	expect := fixture.Module
	err := db.StoreModule(ctx, expect)
	if err != nil {
		t.Fatal(err)
	}

	// Find.
	got, err := db.FindModuleByUUID(ctx, expect.UUID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expect, got); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}

func TestDBPackage(t *testing.T) {
	db := dbtest.Open(t)

	// Store.
	ctx := context.Background()
	expect := fixture.Package
	err := db.StorePackage(ctx, expect)
	if err != nil {
		t.Fatal(err)
	}

	// Find.
	got, err := db.FindPackageByUUID(ctx, expect.UUID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expect, got); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}

func TestDBBenchmark(t *testing.T) {
	db := dbtest.Open(t)

	// Store.
	ctx := context.Background()
	expect := fixture.Benchmark
	err := db.StoreBenchmark(ctx, expect)
	if err != nil {
		t.Fatal(err)
	}

	// Find.
	got, err := db.FindBenchmarkByUUID(ctx, expect.UUID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expect, got); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}

func TestDBDataFile(t *testing.T) {
	db := dbtest.Open(t)

	// Store.
	ctx := context.Background()
	expect := fixture.DataFile
	err := db.StoreDataFile(ctx, expect)
	if err != nil {
		t.Fatal(err)
	}

	// Find.
	got, err := db.FindDataFileByUUID(ctx, expect.UUID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expect, got); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}

func TestDBProperties(t *testing.T) {
	db := dbtest.Open(t)

	cases := []entity.Properties{
		nil,
		map[string]string{},
		map[string]string{
			"a": "1",
			"b": "2",
		},
	}

	for _, expect := range cases {
		t.Logf("uuid=%s", expect.UUID())
		t.Logf("fields=%s", expect)

		// Store.
		ctx := context.Background()
		err := db.StoreProperties(ctx, expect)
		if err != nil {
			t.Fatal(err)
		}

		// Find.
		got, err := db.FindPropertiesByUUID(ctx, expect.UUID())
		if err != nil {
			t.Fatal(err)
		}

		if diff := cmp.Diff(expect, got); diff != "" {
			t.Errorf("mismatch\n%s", diff)
		}
	}
}

func TestDBResult(t *testing.T) {
	db := dbtest.Open(t)

	// Store.
	ctx := context.Background()
	expect := fixture.Result
	err := db.StoreResult(ctx, expect)
	if err != nil {
		t.Fatal(err)
	}

	// Find.
	got, err := db.FindResultByUUID(ctx, expect.UUID())
	if err != nil {
		t.Fatal(err)
	}

	if diff := cmp.Diff(expect, got); diff != "" {
		t.Errorf("mismatch\n%s", diff)
	}
}
