package db

import (
	"context"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math"

	"github.com/mmcloughlin/goperf/app/change"
	"github.com/mmcloughlin/goperf/app/db/internal/db"
	"github.com/mmcloughlin/goperf/app/entity"
)

// StoreChangesBatch writes the given changes to the database in a single batch.
// Does not write any dependent objects.
func (d *DB) StoreChangesBatch(ctx context.Context, cs []*entity.Change) error {
	return d.tx(ctx, func(tx *sql.Tx) error {
		return d.storeChangesBatch(ctx, tx, cs)
	})
}

// ReplaceChanges transactionally deletes changes in a range and inserts supplied changes.
func (d *DB) ReplaceChanges(ctx context.Context, r entity.CommitIndexRange, cs []*entity.Change) error {
	return d.tx(ctx, func(tx *sql.Tx) error {
		if err := d.q.WithTx(tx).DeleteChangesCommitRange(ctx, db.DeleteChangesCommitRangeParams{
			CommitIndexMin: int32(r.Min),
			CommitIndexMax: int32(r.Max),
		}); err != nil {
			return err
		}

		return d.storeChangesBatch(ctx, tx, cs)
	})
}

func (d *DB) storeChangesBatch(ctx context.Context, tx *sql.Tx, cs []*entity.Change) error {
	fields := []string{
		"benchmark_uuid",
		"environment_uuid",
		"commit_index",
		"effect_size",
		"pre_n",
		"pre_mean",
		"pre_stddev",
		"post_n",
		"post_mean",
		"post_stddev",
	}
	values := []interface{}{}
	for _, c := range cs {
		values = append(values,
			c.BenchmarkUUID,
			c.EnvironmentUUID,
			c.CommitIndex,
			c.EffectSize,
			c.Pre.N,
			c.Pre.Mean,
			c.Pre.Stddev(),
			c.Post.N,
			c.Post.Mean,
			c.Post.Stddev(),
		)
	}
	return d.insert(ctx, tx, "changes", fields, values)
}

// BuildChangesRanked derives the ranked changes table from the changes table.
func (d *DB) BuildChangesRanked(ctx context.Context) error {
	return d.txq(ctx, func(q *db.Queries) error {
		return q.BuildChangesRanked(ctx)
	})
}

// ChangeFilter specifies thresholds for change listings.
type ChangeFilter struct {
	MinEffectSize             float64
	MaxRankByEffectSize       int
	MaxRankByAbsPercentChange int
}

// ListChangeSummariesForCommitIndex returns changes at a specific commit.
func (d *DB) ListChangeSummariesForCommitIndex(ctx context.Context, idx int, filter ChangeFilter) ([]*entity.ChangeSummary, error) {
	return d.ListChangeSummaries(ctx, entity.SingleCommitIndexRange(idx), filter)
}

// ListChangeSummaries returns changes with associated metadata.
func (d *DB) ListChangeSummaries(ctx context.Context, r entity.CommitIndexRange, filter ChangeFilter) ([]*entity.ChangeSummary, error) {
	var cs []*entity.ChangeSummary
	err := d.txq(ctx, func(q *db.Queries) error {
		var err error
		cs, err = listChangeSummaries(ctx, q, r, filter)
		return err
	})
	return cs, err
}

func listChangeSummaries(ctx context.Context, q *db.Queries, r entity.CommitIndexRange, filter ChangeFilter) ([]*entity.ChangeSummary, error) {
	rows, err := q.ChangeSummaries(ctx, db.ChangeSummariesParams{
		EffectSizeMin:             filter.MinEffectSize,
		CommitIndexMin:            int32(r.Min),
		CommitIndexMax:            int32(r.Max),
		RankByEffectSizeMax:       zeroToMax32(filter.MaxRankByEffectSize),
		RankByAbsPercentChangeMax: zeroToMax32(filter.MaxRankByAbsPercentChange),
	})
	if err != nil {
		return nil, err
	}

	cs := make([]*entity.ChangeSummary, len(rows))
	for i, row := range rows {
		params := map[string]string{}
		if err := json.Unmarshal(row.Parameters, &params); err != nil {
			return nil, fmt.Errorf("decode parameters: %w", err)
		}

		cs[i] = &entity.ChangeSummary{
			Benchmark: &entity.Benchmark{
				Package: &entity.Package{
					Module: &entity.Module{
						Path:    row.Path,
						Version: row.Version,
					},
					RelativePath: row.RelativePath,
				},
				FullName:   row.FullName,
				Name:       row.Name,
				Parameters: params,
				Unit:       row.Unit,
			},
			EnvironmentUUID: row.EnvironmentUUID,
			CommitSHA:       hex.EncodeToString(row.CommitSHA),
			CommitSubject:   row.CommitSubject,
			Change: change.Change{
				CommitIndex: int(row.CommitIndex),
				EffectSize:  row.EffectSize,
				Pre: change.Stats{
					N:        int(row.PreN),
					Mean:     row.PreMean,
					Variance: row.PreStddev * row.PreStddev,
				},
				Post: change.Stats{
					N:        int(row.PostN),
					Mean:     row.PostMean,
					Variance: row.PostStddev * row.PostStddev,
				},
			},
		}
	}

	return cs, nil
}

func zeroToMax32(x int) int32 {
	if x == 0 {
		return math.MaxInt32
	}
	return int32(x)
}
