package api

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"net/http"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/andrieee44/countmein/gen/users/v2"
	"github.com/andrieee44/countmein/gen/users/v2/usersv2connect"
	"github.com/andrieee44/countmein/store/v2"
)

type actorKey struct{}

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

func withActor(ctx context.Context, actor UserActor) context.Context {
	return context.WithValue(ctx, actorKey{}, actor)
}

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
		sessionHash []byte
		err         error
	)

	sessionHash, err = u.sessionHashFromHeader(req.Header())
	if err != nil {
		return nil, err
	}

	err = store.New(u.db).RevokeUserSession(ctx, sessionHash)
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
	userID int64,
) (string, error) {
	var (
		sessionToken []byte
		sessionHash  [32]byte
		err          error
	)

	sessionToken = make([]byte, 32)
	_, err = rand.Read(sessionToken)
	if err != nil {
		return "", err
	}

	sessionHash = sha256.Sum256(sessionToken)

	err = store.New(u.db).CreateUserSession(
		ctx,
		store.CreateUserSessionParams{
			SessionHash: sessionHash[:],
			UserID:      userID,
			TtlSeconds:  int64(24 * time.Hour / time.Second),
		},
	)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(sessionToken), nil
}

func (u *UserSessionService) AuthInterceptor() connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			var (
				sessionHash []byte
				tx          *sql.Tx
				row         store.GetUserSessionRow
				err         error
			)

			sessionHash, err = u.sessionHashFromHeader(req.Header())
			if sessionHash == nil && err == nil {
				return next(ctx, req)
			}

			tx, err = u.db.BeginTx(ctx, nil)
			if err != nil {
				return nil, err
			}

			defer tx.Rollback()

			row, err = store.New(tx).GetUserSession(ctx, sessionHash)
			if err != nil {
				return nil, err
			}

			if row.DBTime.After(row.ExpiresAt) {
				err = store.New(tx).RevokeUserSession(ctx, sessionHash)
				if err != nil {
					return nil, err
				}

				return nil, ErrSessionExpired
			}

			err = tx.Commit()
			if err != nil {
				return nil, err
			}

			return next(withActor(ctx, UserActor{
				UserID: row.UserID,
			}), req)
		}
	}
}

func (UserSessionService) sessionHashFromHeader(
	header http.Header,
) ([]byte, error) {
	var (
		headerToken  string
		sessionToken []byte
		sessionHash  [32]byte
		err          error
	)

	headerToken = header.Get("Authorization")
	headerToken = strings.TrimPrefix(headerToken, "Bearer ")

	if headerToken == "" {
		return nil, nil
	}

	sessionToken, err = base64.StdEncoding.DecodeString(headerToken)
	if err != nil {
		return nil, err
	}

	sessionHash = sha256.Sum256(sessionToken)

	return sessionHash[:], nil
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
