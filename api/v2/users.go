package api

import (
	"context"
	"database/sql"
	"net/http"
	"net/mail"

	"connectrpc.com/connect"
	"github.com/andrieee44/countmein/gen/users/v2"
	"github.com/andrieee44/countmein/gen/users/v2/usersv2connect"
	"github.com/andrieee44/countmein/store/v2"
)

type UserActor struct {
	UserID int64
}

type UserService struct {
	db             *sql.DB
	sessionService *UserSessionService
}

func NewUserService(
	db *sql.DB,
	sessionService *UserSessionService,
) *UserService {
	return &UserService{
		db:             db,
		sessionService: sessionService,
	}
}

func (u *UserService) CreateUser(
	ctx context.Context,
	req *connect.Request[usersv2.CreateUserRequest],
) (*connect.Response[usersv2.CreateUserResponse], error) {
	var (
		passwordHash []byte
		userID       int64
		sessionToken string
		err          error
	)

	err = u.validateEmail(req.Msg.Email)
	if err != nil {
		return nil, err
	}

	passwordHash, err = Hash(req.Msg.Password)
	if err != nil {
		return nil, err
	}

	userID, err = store.New(u.db).CreateUser(ctx, store.CreateUserParams{
		Email:        req.Msg.Email,
		FirstName:    req.Msg.FirstName,
		LastName:     req.Msg.LastName,
		PasswordHash: passwordHash,
		MiddleName:   fromPtr(req.Msg.MiddleName),
	})
	if err != nil {
		return nil, err
	}

	sessionToken, err = u.sessionService.createSession(ctx, userID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.CreateUserResponse{
		UserId:       userID,
		SessionToken: sessionToken,
	}), nil
}

func (u *UserService) LoginUser(
	ctx context.Context,
	req *connect.Request[usersv2.LoginUserRequest],
) (*connect.Response[usersv2.LoginUserResponse], error) {
	var (
		row          store.GetLoginUserRow
		sessionToken string
		ok           bool
		err          error
	)

	row, err = store.New(u.db).GetLoginUser(ctx, req.Msg.Email)
	if err != nil {
		return nil, err
	}

	ok, err = ComparePlainToHash(req.Msg.Password, row.PasswordHash)
	if err != nil {
		return nil, err
	}

	if !ok {
		return nil, ErrInvalidPassword
	}

	sessionToken, err = u.sessionService.createSession(ctx, row.UserID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.LoginUserResponse{
		SessionToken: sessionToken,
	}), nil
}

func (u *UserService) GetUser(
	ctx context.Context,
	req *connect.Request[usersv2.GetUserRequest],
) (*connect.Response[usersv2.GetUserResponse], error) {
	var (
		row store.GetUserRow
		err error
	)

	row, err = store.New(u.db).GetUser(ctx, req.Msg.UserId)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.GetUserResponse{
		Email:      row.Email,
		FirstName:  row.FirstName,
		LastName:   row.LastName,
		MiddleName: toPtr(row.MiddleName),
	}), nil
}

func (u *UserService) UpdateUser(
	ctx context.Context,
	req *connect.Request[usersv2.UpdateUserRequest],
) (*connect.Response[usersv2.UpdateUserResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = store.New(u.db).UpdateUser(ctx, store.UpdateUserParams{
		ActorUserID: actor.UserID,
		FirstName:   fromPtr(req.Msg.FirstName),
		LastName:    fromPtr(req.Msg.LastName),
		MiddleName:  fromPtr(req.Msg.MiddleName),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.UpdateUserResponse{}), nil
}

func (u *UserService) UpdateLoginUser(
	ctx context.Context,
	req *connect.Request[usersv2.UpdateLoginUserRequest],
) (*connect.Response[usersv2.UpdateLoginUserResponse], error) {
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

	params.ActorUserID = actor.UserID

	if req.Msg.Password != nil {
		hash, err = Hash(*req.Msg.Password)
		if err != nil {
			return nil, err
		}

		params.PasswordHash = fromPtr(&hash)
	}

	if req.Msg.Email != nil {
		err = u.validateEmail(*req.Msg.Email)
		if err != nil {
			return nil, err
		}

		params.Email = fromPtr(req.Msg.Email)
	}

	err = store.New(u.db).UpdateLoginUser(ctx, params)
	if err != nil {
		return nil, err
	}

	_, err = u.sessionService.RevokeAllSessions(
		context.WithoutCancel(ctx),
		connect.NewRequest(&usersv2.RevokeAllSessionsRequest{}),
	)
	if err != nil {
		panic(err)
	}

	return connect.NewResponse(&usersv2.UpdateLoginUserResponse{}), nil
}

func (u *UserService) DeleteUser(
	ctx context.Context,
	req *connect.Request[usersv2.DeleteUserRequest],
) (*connect.Response[usersv2.DeleteUserResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = store.New(u.db).DeleteUser(ctx, actor.UserID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.DeleteUserResponse{}), nil
}

func (UserService) validateEmail(email string) error {
	var (
		addr *mail.Address
		err  error
	)

	addr, err = mail.ParseAddress(email)
	if err != nil {
		return err
	}

	if addr.Address != email {
		return ErrDisplayNameNotAllowed
	}

	return nil
}

func UserHandler(
	db *sql.DB,
	sessionService *UserSessionService,
	opts ...connect.HandlerOption,
) RPCHandlerFn {
	return func() (string, http.Handler) {
		return usersv2connect.NewUserServiceHandler(
			NewUserService(db, sessionService),
			opts...,
		)
	}
}
