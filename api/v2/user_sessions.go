package api

import (
	"context"
	"database/sql"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/andrieee44/countmein/gen/users/v2"
	"github.com/andrieee44/countmein/gen/users/v2/usersv2connect"
	"github.com/andrieee44/countmein/store/v2"
	"github.com/google/uuid"
)

type UserSessionService struct {
	db *sql.DB
}

func NewUserSessionService(db *sql.DB) *UserSessionService {
	return &UserSessionService{
		db: db,
	}
}

func (u *UserSessionService) GetSessionUserID(
	ctx context.Context,
	req *connect.Request[usersv2.GetSessionUserIDRequest],
) (*connect.Response[usersv2.GetSessionUserIDResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.GetSessionUserIDResponse{
		UserId: actor.UserID,
	}), nil
}

func (u *UserSessionService) RevokeSession(
	ctx context.Context,
	req *connect.Request[usersv2.RevokeSessionRequest],
) (*connect.Response[usersv2.RevokeSessionResponse], error) {
	var (
		id  uuid.UUID
		err error
	)

	id, err = u.tokenIDFromHeader(req.Header())
	if err != nil {
		return nil, err
	}

	err = store.New(u.db).RevokeUserSession(ctx, id[:])
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.RevokeSessionResponse{}), nil
}

func (u *UserSessionService) RevokeAllSessions(
	ctx context.Context,
	req *connect.Request[usersv2.RevokeAllSessionsRequest],
) (*connect.Response[usersv2.RevokeAllSessionsResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = store.New(u.db).RevokeAllUserSessions(ctx, actor.UserID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.RevokeAllSessionsResponse{}), nil
}

func (u *UserSessionService) createSession(
	ctx context.Context,
	userID int32,
) (uuid.UUID, error) {
	var (
		sessionID uuid.UUID
		err       error
	)

	sessionID, err = uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	err = store.New(u.db).CreateUserSession(
		ctx,
		store.CreateUserSessionParams{
			ID:         sessionID[:],
			UserID:     userID,
			TtlSeconds: int64(24 * time.Hour / time.Second),
		},
	)
	if err != nil {
		return uuid.Nil, err
	}

	return sessionID, nil
}

func (u *UserSessionService) AuthInterceptor() connect.UnaryInterceptorFunc {
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

			id, err = u.tokenIDFromHeader(req.Header())
			if id == uuid.Nil && err == nil {
				return next(ctx, req)
			}

			if err != nil {
				return nil, err
			}

			row, err = store.New(u.db).GetUserSession(ctx, id[:])
			if err != nil {
				return nil, err
			}

			if row.DBTime.After(row.ExpiresAt) {
				err = store.New(u.db).RevokeUserSession(ctx, id[:])
				if err != nil {
					return nil, err
				}

				return nil, ErrSessionExpired
			}

			return next(WithActor(ctx, UserActor{
				UserID: row.UserID,
			}), req)
		}
	}
}

func (UserSessionService) tokenIDFromHeader(
	header http.Header,
) (uuid.UUID, error) {
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

func UserSessionHandler(
	db *sql.DB,
	opts ...connect.HandlerOption,
) RPCHandlerFn {
	return func() (string, http.Handler) {
		return usersv2connect.NewUserSessionServiceHandler(
			NewUserSessionService(db),
			opts...,
		)
	}
}
