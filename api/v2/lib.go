package api

import (
	"context"
	"database/sql"
	"net/http"
)

type RPCHandlerFn func() (string, http.Handler)

type actorKey struct{}

func WithActor(ctx context.Context, actor UserActor) context.Context {
	return context.WithValue(ctx, actorKey{}, actor)
}

func ActorFromContext(ctx context.Context) (UserActor, error) {
	var (
		actor UserActor
		ok    bool
	)

	actor, ok = ctx.Value(actorKey{}).(UserActor)
	if !ok {
		return UserActor{}, ErrAuthRequired
	}

	return actor, nil
}

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
