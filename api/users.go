package api

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/mail"
	"time"

	"connectrpc.com/connect"
	usersv1 "github.com/andrieee44/countmein/gen/users/v1"
	"github.com/andrieee44/countmein/gen/users/v1/usersv1connect"
	"github.com/andrieee44/countmein/store"
	"github.com/google/uuid"
)

type UserActor struct {
	ID    int32
	Email string
}

type UserService struct {
	db *sql.DB
}

func NewUserService(db *sql.DB) *UserService {
	return &UserService{
		db: db,
	}
}

func (u *UserService) Create(
	ctx context.Context,
	req *connect.Request[usersv1.CreateRequest],
) (*connect.Response[usersv1.CreateResponse], error) {
	var (
		hash      []byte
		result    sql.Result
		userID    int64
		sessionID uuid.UUID
		err       error
	)

	err = validateEmail(req.Msg.Email)
	if err != nil {
		return nil, err
	}

	hash, err = Hash(req.Msg.Password)
	if err != nil {
		return nil, err
	}

	result, err = store.New(u.db).CreateUser(ctx, store.CreateUserParams{
		Email:        req.Msg.Email,
		FirstName:    req.Msg.FirstName,
		LastName:     req.Msg.LastName,
		PasswordHash: hash,
		MiddleName:   fromPtr(req.Msg.MiddleName),
	})
	if err != nil {
		return nil, err
	}

	userID, err = result.LastInsertId()
	if err != nil {
		return nil, err
	}

	sessionID, err = createSession(ctx, u.db, int32(userID), req.Msg.Email)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv1.CreateResponse{
		SessionId: sessionID.String(),
	}), nil

}

func (u *UserService) Login(
	ctx context.Context,
	req *connect.Request[usersv1.LoginRequest],
) (*connect.Response[usersv1.LoginResponse], error) {
	var (
		tx        *sql.Tx
		row       store.GetLoginUserRow
		sessionID uuid.UUID
		ok        bool
		err       error
	)

	tx, err = u.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	row, err = store.New(tx).GetLoginUser(ctx, req.Msg.Email)
	if err != nil {
		return nil, err
	}

	ok, err = ComparePlainToHash(req.Msg.Password, row.PasswordHash)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, errors.New("password verification failed")
	}

	sessionID, err = createSession(ctx, tx, row.ID, req.Msg.Email)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv1.LoginResponse{
		SessionId: sessionID.String(),
	}), nil
}

func (u *UserService) Get(
	ctx context.Context,
	req *connect.Request[usersv1.GetRequest],
) (*connect.Response[usersv1.GetResponse], error) {
	var (
		row store.GetUserRow
		err error
	)

	row, err = store.New(u.db).GetUser(ctx, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv1.GetResponse{
		Email:      row.Email,
		FirstName:  row.FirstName,
		LastName:   row.LastName,
		MiddleName: toPtr(row.MiddleName),
	}), nil

}

func (u *UserService) Update(
	ctx context.Context,
	req *connect.Request[usersv1.UpdateRequest],
) (*connect.Response[usersv1.UpdateResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = store.New(u.db).UpdateUser(ctx, store.UpdateUserParams{
		ID:         actor.ID,
		FirstName:  fromPtr(req.Msg.FirstName),
		LastName:   fromPtr(req.Msg.LastName),
		MiddleName: fromPtr(req.Msg.MiddleName),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv1.UpdateResponse{}), nil
}

func (u *UserService) UpdateLogin(
	ctx context.Context,
	req *connect.Request[usersv1.UpdateLoginRequest],
) (*connect.Response[usersv1.UpdateLoginResponse], error) {
	var (
		actor  UserActor
		params store.UpdateLoginUserParams
		hash   []byte
		err    error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	params.ID = actor.ID

	if req.Msg.Password != nil {
		hash, err = Hash(*req.Msg.Password)
		if err != nil {
			return nil, err
		}

		params.PasswordHash = fromPtr(&hash)
	}

	if req.Msg.Email != nil {
		err = validateEmail(*req.Msg.Email)
		if err != nil {
			return nil, err
		}

		params.Email = fromPtr(req.Msg.Email)
	}

	err = store.New(u.db).UpdateLoginUser(ctx, params)
	if err != nil {
		return nil, err
	}

	err = store.New(u.db).UpdateUserSession(ctx, actor.ID)
	if err != nil {
		err = store.New(u.db).RevokeAllUserSession(ctx, actor.ID)
		if err != nil {
			panic(err)
		}

		return nil, err
	}

	return connect.NewResponse(&usersv1.UpdateLoginResponse{}), nil
}

func (u *UserService) Delete(
	ctx context.Context,
	req *connect.Request[usersv1.DeleteRequest],
) (*connect.Response[usersv1.DeleteResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = store.New(u.db).DeleteUser(ctx, actor.ID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv1.DeleteResponse{}), nil
}

func (u *UserService) Revoke(
	ctx context.Context,
	req *connect.Request[usersv1.RevokeRequest],
) (*connect.Response[usersv1.RevokeResponse], error) {
	var (
		id  uuid.UUID
		err error
	)

	id, err = tokenIDFromHeader(req.Header())
	if err != nil {
		return nil, err
	}

	err = store.New(u.db).RevokeUserSession(ctx, id[:])
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv1.RevokeResponse{}), nil
}

func (u *UserService) RevokeAll(
	ctx context.Context,
	req *connect.Request[usersv1.RevokeAllRequest],
) (*connect.Response[usersv1.RevokeAllResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = store.New(u.db).RevokeAllUserSession(ctx, actor.ID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv1.RevokeAllResponse{}), nil
}

func NewUserHandler(
	db *sql.DB,
	opts ...connect.HandlerOption,
) (string, http.Handler) {
	return usersv1connect.NewUserServiceHandler(
		NewUserService(db),
		opts...,
	)
}

func validateEmail(email string) error {
	var (
		addr *mail.Address
		err  error
	)

	addr, err = mail.ParseAddress(email)
	if err != nil {
		return err
	}

	if addr.Address != email {
		return errors.New("display names not allowed")
	}

	return nil
}

func createSession(
	ctx context.Context,
	db store.DBTX,
	userID int32,
	email string,
) (uuid.UUID, error) {
	var (
		sessionID uuid.UUID
		current   time.Time
		err       error
	)

	sessionID, err = uuid.NewRandom()
	if err != nil {
		return uuid.Nil, err
	}

	current, err = store.New(db).GetCurrentTime(ctx)
	if err != nil {
		return uuid.Nil, err
	}

	err = store.New(db).CreateUserSession(
		ctx,
		store.CreateUserSessionParams{
			ID:        sessionID[:],
			UserID:    userID,
			Email:     email,
			ExpiresAt: current.Add(24 * time.Hour),
		},
	)
	if err != nil {
		return uuid.Nil, err
	}

	return sessionID, nil
}
