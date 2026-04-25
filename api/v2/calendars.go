package api

import (
	"bytes"
	"context"
	"database/sql"
	_ "embed"
	"net/http"

	"connectrpc.com/connect"
	"github.com/andrieee44/countmein/gen/calendars/v2"
	"github.com/andrieee44/countmein/gen/calendars/v2/calendarsv2connect"
	"github.com/andrieee44/countmein/store/v2"
	ics "github.com/arran4/golang-ical"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CalendarService struct {
	db              *sql.DB
	_AES_SECRET_KEY []byte
}

func NewCalendarService(
	db *sql.DB,
	_AES_SECRET_KEY []byte,
) *CalendarService {
	return &CalendarService{
		db:              db,
		_AES_SECRET_KEY: _AES_SECRET_KEY,
	}
}

func (c *CalendarService) CreateCalendar(
	ctx context.Context,
	req *connect.Request[calendarsv2.CreateCalendarRequest],
) (*connect.Response[calendarsv2.CreateCalendarResponse], error) {
	var (
		actor      UserActor
		tx         *sql.Tx
		calendarID int64
		err        error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	_, err = ics.ParseCalendar(bytes.NewReader(req.Msg.Ical))
	if err != nil {
		return nil, err
	}

	tx, err = c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	err = store.New(tx).CreateCalendar(ctx, store.CreateCalendarParams{
		ActorUserID:  actor.UserID,
		Name:         req.Msg.Name,
		Ical:         req.Msg.Ical,
		Description:  req.Msg.Description,
		AesSecretKey: c._AES_SECRET_KEY,
	})
	if err != nil {
		return nil, err
	}

	err = tx.QueryRowContext(ctx, "SELECT @out_calendar_id").Scan(
		&calendarID,
	)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv2.CreateCalendarResponse{
		CalendarId: calendarID,
	}), nil
}

func (c *CalendarService) GetCalendar(
	ctx context.Context,
	req *connect.Request[calendarsv2.GetCalendarRequest],
) (*connect.Response[calendarsv2.GetCalendarResponse], error) {
	var (
		actor UserActor
		row   store.GetCalendarRow
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	row, err = store.New(c.db).GetCalendar(ctx, store.GetCalendarParams{
		ActorUserID:  actor.UserID,
		CalendarID:   req.Msg.CalendarId,
		AesSecretKey: string(c._AES_SECRET_KEY),
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv2.GetCalendarResponse{
		OwnerUserId: row.OwnerUserID,
		Name:        row.Name,
		Ical:        []byte(row.Ical),
		UpdatedAt:   timestamppb.New(row.UpdatedAt),
		Description: toPtr(row.Description),
	}), nil
}

func (c *CalendarService) GetCalendarWrites(
	ctx context.Context,
	req *connect.Request[calendarsv2.GetCalendarWritesRequest],
) (*connect.Response[calendarsv2.GetCalendarWritesResponse], error) {
	var (
		actor    UserActor
		eventIDs []int64
		err      error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	eventIDs, err = store.New(c.db).GetCalendarWrites(
		ctx,
		store.GetCalendarWritesParams{
			ActorUserID: actor.UserID,
			CalendarID:  req.Msg.CalendarId,
		},
	)
	if err != nil {
		return nil, err
	}

	if len(eventIDs) == 0 {
		return nil, ErrNotFoundOrDenied
	}

	return connect.NewResponse(&calendarsv2.GetCalendarWritesResponse{
		CalendarWriteEventIds: eventIDs,
	}), nil
}

func (c *CalendarService) WriteCalendar(
	ctx context.Context,
	req *connect.Request[calendarsv2.WriteCalendarRequest],
) (*connect.Response[calendarsv2.WriteCalendarResponse], error) {
	var (
		actor      UserActor
		calendarID int64
		err        error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	_, err = ics.ParseCalendar(bytes.NewReader(req.Msg.Ical))
	if err != nil {
		return nil, err
	}

	calendarID, err = store.New(c.db).WriteCalendar(
		ctx,
		store.WriteCalendarParams{
			ActorUserID:  actor.UserID,
			ActorUserId2: fromPtr(&actor.UserID),
			CalendarID:   req.Msg.CalendarId,
			Ical:         string(req.Msg.Ical),
			AesSecretKey: string(c._AES_SECRET_KEY),
		},
	)
	if err != nil {
		return nil, err
	}

	if calendarID == 0 {
		return nil, ErrNotFoundOrDenied
	}

	return connect.NewResponse(&calendarsv2.WriteCalendarResponse{}), nil
}

func (c *CalendarService) UpdateCalendarMetadata(
	ctx context.Context,
	req *connect.Request[calendarsv2.UpdateCalendarMetadataRequest],
) (*connect.Response[calendarsv2.UpdateCalendarMetadataResponse], error) {
	var (
		actor UserActor
		n     int64
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	n, err = store.New(c.db).UpdateCalendarMetadata(
		ctx,
		store.UpdateCalendarMetadataParams{
			ActorUserID: actor.UserID,
			CalendarID:  req.Msg.CalendarId,
			Name:        fromPtr(req.Msg.Name),
			Description: fromPtr(req.Msg.Description),
		},
	)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, ErrNotFoundOrDenied
	}

	return connect.NewResponse(
		&calendarsv2.UpdateCalendarMetadataResponse{},
	), nil
}

func (c *CalendarService) DeleteCalendar(
	ctx context.Context,
	req *connect.Request[calendarsv2.DeleteCalendarRequest],
) (*connect.Response[calendarsv2.DeleteCalendarResponse], error) {
	var (
		actor UserActor
		n     int64
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	n, err = store.New(c.db).DeleteCalendar(
		ctx,
		store.DeleteCalendarParams{
			ActorUserID: actor.UserID,
			CalendarID:  req.Msg.CalendarId,
		},
	)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, ErrNotFoundOrDenied
	}

	return connect.NewResponse(
		&calendarsv2.DeleteCalendarResponse{},
	), nil
}

func CalendarHandler(
	db *sql.DB,
	_AES_SECRET_KEY []byte,
	opts ...connect.HandlerOption,
) RPCHandlerFn {
	return func() (string, http.Handler) {
		return calendarsv2connect.NewCalendarServiceHandler(
			NewCalendarService(db, _AES_SECRET_KEY),
			opts...,
		)
	}
}
