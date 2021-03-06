package ingest

import (
	"context"
	"fmt"
	"path"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/db"
	"github.com/mmcloughlin/goperf/app/gcs"
	"github.com/mmcloughlin/goperf/app/ingest"
	"github.com/mmcloughlin/goperf/app/results"
	"github.com/mmcloughlin/goperf/app/service"
)

// Initialization.
var (
	logger   *zap.Logger
	database *db.DB
)

func initialize(ctx context.Context, l *zap.Logger) error {
	var err error

	logger = l

	database, err = service.DB(ctx, logger)
	if err != nil {
		return err
	}

	return nil
}

func init() {
	service.Initialize(initialize)
}

// GCSEvent is the payload of a GCS event.
type GCSEvent struct {
	Bucket string `json:"bucket"`
	Name   string `json:"name"`
}

// Handle GCS event.
func Handle(ctx context.Context, e GCSEvent) error {
	logger.Info("received cloud storage trigger",
		zap.String("bucket", e.Bucket),
		zap.String("name", e.Name),
	)

	// Extract task ID from the object name.
	id, err := uuid.Parse(path.Base(e.Name))
	if err != nil {
		return fmt.Errorf("parse task uuid from name: %w", err)
	}

	// Construct Ingester.
	bucket, err := gcs.New(ctx, e.Bucket)
	if err != nil {
		return err
	}

	loader, err := results.NewLoader(results.WithFilesystem(bucket))
	if err != nil {
		return err
	}

	i := ingest.New(database, loader)
	i.SetLogger(logger)

	// Ingest task.
	if err := i.Task(ctx, id); err != nil {
		return fmt.Errorf("task ingest: %w", err)
	}

	return nil
}
