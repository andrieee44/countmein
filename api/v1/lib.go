package api

import "database/sql"

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
