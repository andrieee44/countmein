package api

import (
	"context"
	"database/sql"
	_ "embed"
	"net/http"

	"connectrpc.com/connect"
	"github.com/andrieee44/countmein/gen/calendars/v2"
	"github.com/andrieee44/countmein/gen/calendars/v2/calendarsv2connect"
	"github.com/andrieee44/countmein/store/v2"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type CalendarWriteService struct {
	db              *sql.DB
	_AES_SECRET_KEY []byte
}

func NewCalendarWriteService(
	db *sql.DB,
	_AES_SECRET_KEY []byte,
) *CalendarWriteService {
	return &CalendarWriteService{
		db:              db,
		_AES_SECRET_KEY: _AES_SECRET_KEY,
	}
}

func (c *CalendarWriteService) GetCalendarWrite(
	ctx context.Context,
	req *connect.Request[calendarsv2.GetCalendarWriteRequest],
) (*connect.Response[calendarsv2.GetCalendarWriteResponse], error) {
	var (
		actor UserActor
		row   store.GetCalendarWriteRow
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	row, err = store.New(c.db).GetCalendarWrite(
		ctx,
		store.GetCalendarWriteParams{
			ActorUserID:          actor.UserID,
			CalendarWriteEventID: req.Msg.CalendarWriteEventId,
			AesSecretKey:         string(c._AES_SECRET_KEY),
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv2.GetCalendarWriteResponse{
		CalendarId:   row.CalendarID,
		WriterUserId: toPtr(row.WriterUserID),
		Ical:         []byte(row.Ical),
		CreatedAt:    timestamppb.New(row.CreatedAt),
	}), nil
}

func CalendarWriteHandler(
	db *sql.DB,
	_AES_SECRET_KEY []byte,
	opts ...connect.HandlerOption,
) RPCHandlerFn {
	return func() (string, http.Handler) {
		return calendarsv2connect.NewCalendarWriteServiceHandler(
			NewCalendarWriteService(db, _AES_SECRET_KEY),
			opts...,
		)
	}
}
