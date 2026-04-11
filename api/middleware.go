package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"connectrpc.com/connect"
	"github.com/andrieee44/countmein/store"
	"github.com/google/uuid"
)

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
		return UserActor{}, errors.New("authentication required")
	}

	return actor, nil
}

func NewErrorInterceptor(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			var (
				res connect.AnyResponse
				err error
			)

			res, err = next(ctx, req)
			if err == nil {
				return res, nil
			}

			logger.Error(
				"RPC Request Failed",
				"procedure", req.Spec().Procedure,
				"protocol", req.Peer().Protocol,
				"error", err,
			)

			return nil, err
		}
	}
}

func NewAuthInterceptor(db store.DBTX) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			var (
				id  uuid.UUID
				row store.GetUserSessionRow
				err error
			)

			id, err = tokenIDFromHeader(req.Header())
			if id == uuid.Nil && err == nil {
				return next(ctx, req)
			}

			if err != nil {
				return nil, err
			}

			row, err = store.New(db).GetUserSession(ctx, id[:])
			if err != nil {
				return nil, err
			}

			if row.DBTime.After(row.ExpiresAt) {
				err = store.New(db).RevokeUserSession(ctx, row.ID)
				if err != nil {
					return nil, err
				}

				return nil, errors.New("session or token expired")
			}

			return next(WithActor(ctx, UserActor{
				ID:    row.UserID,
				Email: row.Email,
			}), req)
		}
	}
}

func tokenIDFromHeader(header http.Header) (uuid.UUID, error) {
	var (
		token string
		id    uuid.UUID
		err   error
	)

	token = header.Get("Authorization")
	token = strings.TrimPrefix(token, "Bearer ")

	if token == "" {
		return uuid.Nil, nil
	}

	id, err = uuid.Parse(token)
	if err != nil {
		return uuid.Nil, err
	}

	return id, nil
}
