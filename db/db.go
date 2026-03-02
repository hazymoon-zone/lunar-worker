package db

import (
	"context"

	"encore.dev/rlog"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"go4.org/syncutil"
)

type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row
}

var secrets struct {
	LunarReminderDatabase string
}

var (
	// once is like sync.Once except it re-arms itself on failure
	once syncutil.Once
	// pool is the successfully created database connection pool,
	// or nil when no such pool has been setup yet.
	pool *pgxpool.Pool
)

// Get returns a database connection pool to the external database.
// It is lazily created on first use.
func GetDatabasePool(ctx context.Context) (*pgxpool.Pool, error) {
	// Attempt to setup the database connection pool if it hasn't
	// already been successfully setup.
	err := once.Do(func() error {
		var err error
		pool, err = pgxpool.Connect(ctx, secrets.LunarReminderDatabase)
		return err
	})
	if err != nil {
		rlog.Error("Error connecting to database", "err", err.Error())
		return nil, err
	}

	return pool, nil
}
