package api

import (
	"context"
	"database/sql"
	"net/http"

	"connectrpc.com/connect"
	"github.com/andrieee44/countmein/gen/users/v2"
	"github.com/andrieee44/countmein/gen/users/v2/usersv2connect"
	"github.com/andrieee44/countmein/store/v2"
)

type UserLabelService struct {
	db *sql.DB
}

func NewUserLabelService(
	db *sql.DB,
) *UserLabelService {
	return &UserLabelService{
		db: db,
	}
}

func (u *UserLabelService) CreateUserLabel(
	ctx context.Context,
	req *connect.Request[usersv2.CreateUserLabelRequest],
) (*connect.Response[usersv2.CreateUserLabelResponse], error) {
	var (
		actor   UserActor
		labelID int64
		err     error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	labelID, err = store.New(u.db).CreateUserLabel(
		ctx,
		store.CreateUserLabelParams{
			ActorUserID: actor.UserID,
			Name:        req.Msg.Name,
			Color:       req.Msg.Color,
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.CreateUserLabelResponse{
		UserLabelId: labelID,
	}), nil
}

func (u *UserLabelService) GetUserLabel(
	ctx context.Context,
	req *connect.Request[usersv2.GetUserLabelRequest],
) (*connect.Response[usersv2.GetUserLabelResponse], error) {
	var (
		actor UserActor
		row   store.GetUserLabelRow
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	row, err = store.New(u.db).GetUserLabel(
		ctx,
		store.GetUserLabelParams{
			ActorUserID: actor.UserID,
			UserLabelID: req.Msg.UserLabelId,
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.GetUserLabelResponse{
		Name:  row.Name,
		Color: row.Color,
	}), nil
}

func (u *UserLabelService) UpdateUserLabel(
	ctx context.Context,
	req *connect.Request[usersv2.UpdateUserLabelRequest],
) (*connect.Response[usersv2.UpdateUserLabelResponse], error) {
	var (
		actor UserActor
		n     int64
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	n, err = store.New(u.db).UpdateUserLabel(
		ctx,
		store.UpdateUserLabelParams{
			ActorUserID: actor.UserID,
			UserLabelID: req.Msg.UserLabelId,
			Name:        fromPtr(req.Msg.Name),
			Color:       fromPtr(req.Msg.Color),
		},
	)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, ErrNotFoundOrDenied
	}

	return connect.NewResponse(&usersv2.UpdateUserLabelResponse{}), nil
}

func (u *UserLabelService) DeleteUserLabel(
	ctx context.Context,
	req *connect.Request[usersv2.DeleteUserLabelRequest],
) (*connect.Response[usersv2.DeleteUserLabelResponse], error) {
	var (
		actor UserActor
		n     int64
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	n, err = store.New(u.db).DeleteUserLabel(
		ctx,
		store.DeleteUserLabelParams{
			ActorUserID: actor.UserID,
			UserLabelID: req.Msg.UserLabelId,
		},
	)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, ErrNotFoundOrDenied
	}

	return connect.NewResponse(&usersv2.DeleteUserLabelResponse{}), nil
}

func (u *UserLabelService) AttachUserLabel(
	ctx context.Context,
	req *connect.Request[usersv2.AttachUserLabelRequest],
) (*connect.Response[usersv2.AttachUserLabelResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = store.New(u.db).AttachUserLabel(
		ctx,
		store.AttachUserLabelParams{
			ActorUserID: actor.UserID,
			UserLabelID: req.Msg.UserLabelId,
			CalendarID:  req.Msg.CalendarId,
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.AttachUserLabelResponse{}), nil
}

func (u *UserLabelService) DetachUserLabel(
	ctx context.Context,
	req *connect.Request[usersv2.DetachUserLabelRequest],
) (*connect.Response[usersv2.DetachUserLabelResponse], error) {
	var (
		actor UserActor
		n     int64
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	n, err = store.New(u.db).DetachUserLabel(
		ctx,
		store.DetachUserLabelParams{
			ActorUserID: actor.UserID,
			UserLabelID: req.Msg.UserLabelId,
			CalendarID:  req.Msg.CalendarId,
		},
	)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, ErrNotFoundOrDenied
	}

	return connect.NewResponse(&usersv2.DetachUserLabelResponse{}), nil
}

func UserLabelHandler(
	db *sql.DB,
	opts ...connect.HandlerOption,
) RPCHandlerFn {
	return func() (string, http.Handler) {
		return usersv2connect.NewUserLabelServiceHandler(
			NewUserLabelService(db),
			opts...,
		)
	}
}
