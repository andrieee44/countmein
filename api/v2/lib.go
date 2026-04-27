package api

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
)

type RPCHandlerFn func() (string, http.Handler)

func fromPtr[T any](v *T) sql.Null[T] {
	if v == nil {
		return sql.Null[T]{}
	}

	return sql.Null[T]{
		V:     *v,
		Valid: true,
	}
}

func toPtr[T any](v sql.Null[T]) *T {
	if !v.Valid {
		return nil
	}

	return &v.V
}

func sproc(
	ctx context.Context,
	db *sql.DB,
	outParam string,
	outPtr any,
	fn func(tx *sql.Tx) error,
) error {
	var (
		tx  *sql.Tx
		err error
	)

	tx, err = db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}

	defer tx.Rollback()

	err = fn(tx)
	if err != nil {
		return err
	}

	err = tx.QueryRowContext(ctx, fmt.Sprintf("SELECT %s", outParam)).Scan(
		outPtr,
	)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
