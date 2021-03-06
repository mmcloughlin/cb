package watch

import (
	"context"
	"io"
	"net/http"

	"go.uber.org/zap"

	"github.com/mmcloughlin/goperf/app/entity"
	"github.com/mmcloughlin/goperf/app/httputil"
	"github.com/mmcloughlin/goperf/app/repo"
	"github.com/mmcloughlin/goperf/app/service"
)

// Initialization.
var (
	logger     *zap.Logger
	repository repo.Repository
	handler    http.Handler
)

func initialize(ctx context.Context, l *zap.Logger) error {
	logger = l

	repository = repo.Go(http.DefaultClient)

	handler = httputil.ErrorHandler{
		Handler: httputil.HandlerFunc(handle),
		Log:     logger,
	}

	return nil
}

func init() {
	service.Initialize(initialize)
}

// Handle HTTP trigger.
func Handle(w http.ResponseWriter, r *http.Request) {
	handler.ServeHTTP(w, r)
}

func handle(w http.ResponseWriter, r *http.Request) error {
	ctx := r.Context()

	// Open database connection.
	d, err := service.DB(ctx, logger)
	if err != nil {
		return err
	}
	defer d.Close()

	// Get most recent commit on master in the database.
	latest, err := d.MostRecentCommitWithRef(ctx, "master")
	if err != nil {
		return err
	}
	logger.Info("found latest commit on master in database", zap.String("sha", latest.SHA))

	// Fetch commits until we get to the latest one.
	commits := []*entity.Commit{}
	start := "master"
	for {
		// Fetch commits.
		logger.Info("git log", zap.String("start", start))
		batch, err := repository.Log(ctx, start)
		if err != nil {
			logger.Error("error fetching recent commits", zap.Error(err))
			return err
		}

		logger.Info("fetched recent commits", zap.Int("num_commits", len(batch)))
		commits = append(commits, batch...)

		// Look to see if we've hit the latest one.
		if containsCommit(commits, latest) {
			break
		}

		// Update log starting point.
		start = commits[len(commits)-1].SHA
	}

	// Store new commits in the database.
	if err := d.StoreCommits(ctx, commits); err != nil {
		return err
	}
	logger.Info("inserted commits", zap.Int("num_commits", len(commits)))

	// Record refs.
	it := repo.FirstParent(repo.CommitsIterator(commits))
	refs := []*entity.CommitRef{}
	for {
		c, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}

		refs = append(refs, &entity.CommitRef{
			SHA: c.SHA,
			Ref: "master",
		})
	}

	if err := d.StoreCommitRefs(ctx, refs); err != nil {
		return err
	}
	logger.Info("recorded commit refs")

	// Rebuild commit positions table.
	if err := d.BuildCommitPositions(ctx); err != nil {
		return err
	}
	logger.Info("built commit positions")

	// Report ok.
	httputil.OK(w)

	return nil
}

func containsCommit(commits []*entity.Commit, target *entity.Commit) bool {
	for _, c := range commits {
		if c.SHA == target.SHA {
			return true
		}
	}
	return false
}
