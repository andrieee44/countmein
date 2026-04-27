package api

import (
	"context"
	"database/sql"
	_ "embed"
	"net/http"

	"connectrpc.com/connect"
	"github.com/andrieee44/countmein/gen/organizations/v2"
	"github.com/andrieee44/countmein/gen/organizations/v2/organizationsv2connect"
	"github.com/andrieee44/countmein/store/v2"
)

type OrganizationCalendarService struct {
	db *sql.DB
}

func NewOrganizationCalendarService(
	db *sql.DB,
) *OrganizationCalendarService {
	return &OrganizationCalendarService{
		db: db,
	}
}

func (o *OrganizationCalendarService) ToggleShareUserCalendar(
	ctx context.Context,
	req *connect.Request[organizationsv2.ToggleShareUserCalendarRequest],
) (
	*connect.Response[organizationsv2.ToggleShareUserCalendarResponse],
	error,
) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = store.New(o.db).ToggleShareUserCalendar(
		ctx,
		store.ToggleShareUserCalendarParams{
			ActorUserID:    actor.UserID,
			OrganizationID: req.Msg.OrganizationId,
			CalendarID:     req.Msg.CalendarId,
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(
		&organizationsv2.ToggleShareUserCalendarResponse{},
	), nil
}

func (o *OrganizationCalendarService) GetOrganizationCalendars(
	ctx context.Context,
	req *connect.Request[organizationsv2.GetOrganizationCalendarsRequest],
) (
	*connect.Response[organizationsv2.GetOrganizationCalendarsResponse],
	error,
) {
	var (
		actor       UserActor
		calendarIDs []int64
		err         error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	calendarIDs, err = store.New(o.db).GetOrganizationCalendars(
		ctx,
		store.GetOrganizationCalendarsParams{
			ActorUserID:    actor.UserID,
			OrganizationID: req.Msg.OrganizationId,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(calendarIDs) == 0 {
		return nil, ErrNotFoundOrDenied
	}

	return connect.NewResponse(
		&organizationsv2.GetOrganizationCalendarsResponse{
			CalendarIds: calendarIDs,
		},
	), nil
}

func OrganizationCalendarHandler(
	db *sql.DB,
	opts ...connect.HandlerOption,
) RPCHandlerFn {
	return func() (string, http.Handler) {
		return organizationsv2connect.NewOrganizationCalendarServiceHandler(
			NewOrganizationCalendarService(db),
			opts...,
		)
	}
}
