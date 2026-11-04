package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

// BackgroundStatementTimeout is the per-statement budget for scheduled
// background work. The request pool applies a short default that suits
// interactive requests; whole-table refreshes and batch publication set this
// budget for their own transaction instead.
const BackgroundStatementTimeout = "120s"

// TxBeginner is satisfied by *pgxpool.Pool, *pgxpool.Conn and pgx.Tx.
type TxBeginner interface {
	Begin(context.Context) (pgx.Tx, error)
}

// RunWithBackgroundBudget runs fn inside a transaction whose statement timeout
// is BackgroundStatementTimeout. SET LOCAL scopes the budget to that
// transaction, so the connection returns to the pool with its request default.
// An error from fn rolls the transaction back and is returned unchanged.
func RunWithBackgroundBudget(ctx context.Context, db TxBeginner, fn func(pgx.Tx) error) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("beginning background transaction: %w", err)
	}
	defer tx.Rollback(ctx)
	if _, err := tx.Exec(ctx, "SET LOCAL statement_timeout = '"+BackgroundStatementTimeout+"'"); err != nil {
		return fmt.Errorf("setting background statement budget: %w", err)
	}
	if err := fn(tx); err != nil {
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("committing background transaction: %w", err)
	}
	return nil
}
