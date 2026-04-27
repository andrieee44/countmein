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

type UserProfileService struct {
	db *sql.DB
}

func NewUserProfileService(
	db *sql.DB,
) *UserProfileService {
	return &UserProfileService{
		db: db,
	}
}

func (u *UserProfileService) GetUserOwnedCalendars(
	ctx context.Context,
	req *connect.Request[usersv2.GetUserOwnedCalendarsRequest],
) (*connect.Response[usersv2.GetUserOwnedCalendarsResponse], error) {
	var (
		actor       UserActor
		calendarIDs []int64
		err         error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	calendarIDs, err = store.New(u.db).GetUserOwnedCalendars(
		ctx,
		actor.UserID,
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.GetUserOwnedCalendarsResponse{
		CalendarIds: calendarIDs,
	}), nil
}

func (u *UserProfileService) GetUserOrganizations(
	ctx context.Context,
	req *connect.Request[usersv2.GetUserOrganizationsRequest],
) (*connect.Response[usersv2.GetUserOrganizationsResponse], error) {
	var (
		actor           UserActor
		organizationIDs []int64
		err             error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	organizationIDs, err = store.New(u.db).GetUserOrganizations(
		ctx,
		actor.UserID,
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.GetUserOrganizationsResponse{
		OrganizationIds: organizationIDs,
	}), nil
}

func (u *UserProfileService) GetUserLabels(
	ctx context.Context,
	req *connect.Request[usersv2.GetUserLabelsRequest],
) (*connect.Response[usersv2.GetUserLabelsResponse], error) {
	var (
		actor    UserActor
		labelIDs []int64
		err      error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	labelIDs, err = store.New(u.db).GetUserLabels(
		ctx,
		actor.UserID,
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.GetUserLabelsResponse{
		UserLabelIds: labelIDs,
	}), nil
}

func (u *UserProfileService) GetUserCalendarLabels(
	ctx context.Context,
	req *connect.Request[usersv2.GetUserCalendarLabelsRequest],
) (*connect.Response[usersv2.GetUserCalendarLabelsResponse], error) {
	var (
		actor    UserActor
		labelIDs []int64
		err      error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	labelIDs, err = store.New(u.db).GetUserCalendarLabels(
		ctx,
		store.GetUserCalendarLabelsParams{
			CalendarID:  req.Msg.CalendarId,
			ActorUserID: actor.UserID,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(labelIDs) == 0 {
		return nil, ErrNotFoundOrDenied
	}

	return connect.NewResponse(&usersv2.GetUserCalendarLabelsResponse{
		UserLabelIds: labelIDs,
	}), nil
}

func (u *UserProfileService) GetUserEmail(
	ctx context.Context,
	req *connect.Request[usersv2.GetUserEmailRequest],
) (*connect.Response[usersv2.GetUserEmailResponse], error) {
	var (
		actor UserActor
		email string
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	email, err = store.New(u.db).GetUserEmail(ctx, actor.UserID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&usersv2.GetUserEmailResponse{
		Email: email,
	}), nil
}

func UserProfileHandler(
	db *sql.DB,
	opts ...connect.HandlerOption,
) RPCHandlerFn {
	return func() (string, http.Handler) {
		return usersv2connect.NewUserProfileServiceHandler(
			NewUserProfileService(db),
			opts...,
		)
	}
}
