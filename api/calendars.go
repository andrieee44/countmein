package api

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"connectrpc.com/connect"
	calendarsv1 "github.com/andrieee44/countmein/gen/calendars/v1"
	"github.com/andrieee44/countmein/gen/calendars/v1/calendarsv1connect"
	"github.com/andrieee44/countmein/store"
	ics "github.com/arran4/golang-ical"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var reHex *regexp.Regexp = regexp.MustCompile("^[0-9A-Fa-f]{6}$")

type CalendarService struct {
	db *sql.DB
}

func NewCalendarService(db *sql.DB) *CalendarService {
	return &CalendarService{
		db: db,
	}
}

func (c *CalendarService) Create(
	ctx context.Context,
	req *connect.Request[calendarsv1.CreateRequest],
) (*connect.Response[calendarsv1.CreateResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = isValidColor(req.Msg.Color)
	if err != nil {
		return nil, err
	}

	_, err = ics.ParseCalendar(bytes.NewReader(req.Msg.Ical))
	if err != nil {
		return nil, err
	}

	_, err = store.New(c.db).CreateCalendar(
		ctx,
		store.CreateCalendarParams{
			OwnerID:     actor.ID,
			Name:        req.Msg.Name,
			Ical:        req.Msg.Ical,
			MembersOnly: req.Msg.MembersOnly,
			Color:       req.Msg.Color,
			Description: fromPtr(req.Msg.Description),
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.CreateResponse{}), nil
}

func (c *CalendarService) Get(
	ctx context.Context,
	req *connect.Request[calendarsv1.GetRequest],
) (*connect.Response[calendarsv1.GetResponse], error) {
	var (
		actor  UserActor
		banRow store.IsMemberBannedRow
		calRow store.GetCalendarRow
		color  string
		err    error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	banRow, err = store.New(c.db).IsMemberBanned(
		ctx,
		store.IsMemberBannedParams{
			UserID:     actor.ID,
			CalendarID: req.Msg.Id,
		},
	)
	if err == nil {
		if !banRow.ExpiresAt.Valid {
			return nil, fmt.Errorf("banned permanently: %s", banRow.Reason)
		}

		if banRow.DBTime.Before(banRow.ExpiresAt.Time) {
			return nil, fmt.Errorf(
				"banned until %s: %s",
				banRow.ExpiresAt.Time.Format("Jan 2 2006 3:04 PM"),
				banRow.Reason,
			)
		}

		err = store.New(c.db).UnbanCalendarMember(
			ctx,
			store.UnbanCalendarMemberParams{
				UserID:     actor.ID,
				CalendarID: req.Msg.Id,
			},
		)
		if err != nil {
			return nil, err
		}
	} else if !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	calRow, err = store.New(c.db).GetCalendar(ctx, store.GetCalendarParams{
		ID:     req.Msg.Id,
		UserID: actor.ID,
	})
	if err != nil {
		return nil, err
	}

	if calRow.OwnerID == actor.ID {
		color = calRow.Color
	} else {
		color, err = store.New(c.db).GetSubscribedMetadata(
			ctx,
			store.GetSubscribedMetadataParams{
				UserID:     actor.ID,
				CalendarID: req.Msg.Id,
			},
		)
		if err != nil {
			if !errors.Is(err, sql.ErrNoRows) {
				return nil, err
			}

			color = calRow.Color
		}
	}

	return connect.NewResponse(&calendarsv1.GetResponse{
		OwnerId:     calRow.OwnerID,
		Name:        calRow.Name,
		Ical:        calRow.Ical,
		MembersOnly: calRow.MembersOnly,
		Color:       color,
		Description: toPtr(calRow.Description),
	}), nil
}

func (c *CalendarService) GetOwned(
	ctx context.Context,
	req *connect.Request[calendarsv1.GetOwnedRequest],
) (*connect.Response[calendarsv1.GetOwnedResponse], error) {
	var (
		actor UserActor
		ids   []int32
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ids, err = store.New(c.db).GetOwnedCalendars(ctx, actor.ID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.GetOwnedResponse{
		Ids: ids,
	}), nil
}

func (c *CalendarService) GetSubscribed(
	ctx context.Context,
	req *connect.Request[calendarsv1.GetSubscribedRequest],
) (*connect.Response[calendarsv1.GetSubscribedResponse], error) {
	var (
		actor UserActor
		ids   []int32
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ids, err = store.New(c.db).GetSubscribedCalendars(ctx, actor.ID)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.GetSubscribedResponse{
		Ids: ids,
	}), nil
}

func (c *CalendarService) Merge(
	ctx context.Context,
	req *connect.Request[calendarsv1.MergeRequest],
) (*connect.Response[calendarsv1.MergeResponse], error) {
	var (
		tx                  *sql.Tx
		oldICal, mergedICal []byte
		err                 error
	)

	tx, err = c.db.Begin()
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()

	err = isCalendarOwner(ctx, tx, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	oldICal, err = store.New(tx).GetCalendarICal(ctx, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	mergedICal, err = c.mergeICals(oldICal, req.Msg.Ical)
	if err != nil {
		return nil, err
	}

	err = store.New(tx).ReplaceCalendar(ctx, store.ReplaceCalendarParams{
		ID:   req.Msg.Id,
		Ical: mergedICal,
	})
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.MergeResponse{}), nil
}

func (c *CalendarService) Replace(
	ctx context.Context,
	req *connect.Request[calendarsv1.ReplaceRequest],
) (*connect.Response[calendarsv1.ReplaceResponse], error) {
	var err error

	err = isCalendarOwner(ctx, c.db, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	_, err = ics.ParseCalendar(bytes.NewReader(req.Msg.Ical))
	if err != nil {
		return nil, err
	}

	err = store.New(c.db).ReplaceCalendar(ctx, store.ReplaceCalendarParams{
		ID:   req.Msg.Id,
		Ical: req.Msg.Ical,
	})
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.ReplaceResponse{}), nil
}

func (c *CalendarService) UpdateMetadata(
	ctx context.Context,
	req *connect.Request[calendarsv1.UpdateMetadataRequest],
) (*connect.Response[calendarsv1.UpdateMetadataResponse], error) {
	var err error

	err = isCalendarOwner(ctx, c.db, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	err = store.New(c.db).UpdateMetadataCalendar(
		ctx,
		store.UpdateMetadataCalendarParams{
			ID:          req.Msg.Id,
			Name:        fromPtr(req.Msg.Name),
			MembersOnly: fromPtr(req.Msg.MembersOnly),
			Description: fromPtr(req.Msg.Description),
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.UpdateMetadataResponse{}), nil
}

func (c *CalendarService) UpdateSubscribedMetadata(
	ctx context.Context,
	req *connect.Request[calendarsv1.UpdateSubscribedMetadataRequest],
) (*connect.Response[calendarsv1.UpdateSubscribedMetadataResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	if req.Msg.Color != nil {
		err = isValidColor(*req.Msg.Color)
		if err != nil {
			return nil, err
		}
	}

	err = store.New(c.db).UpdateSubscribedMetadata(
		ctx,
		store.UpdateSubscribedMetadataParams{
			UserID:     actor.ID,
			CalendarID: req.Msg.Id,
			Color:      fromPtr(req.Msg.Color),
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(
		&calendarsv1.UpdateSubscribedMetadataResponse{},
	), nil
}

func (c *CalendarService) Delete(
	ctx context.Context,
	req *connect.Request[calendarsv1.DeleteRequest],
) (*connect.Response[calendarsv1.DeleteResponse], error) {
	var err error

	err = isCalendarOwner(ctx, c.db, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	err = store.New(c.db).DeleteCalendar(ctx, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.DeleteResponse{}), nil
}

func (c *CalendarService) CreateCode(
	ctx context.Context,
	req *connect.Request[calendarsv1.CreateCodeRequest],
) (*connect.Response[calendarsv1.CreateCodeResponse], error) {
	var (
		current time.Time
		valid   bool
		err     error
	)

	err = isCalendarOwner(ctx, c.db, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	current, err = store.New(c.db).GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}

	if req.Msg.Ttl != nil {
		current = current.Add(req.Msg.Ttl.AsDuration())
		valid = true
	}

	_, err = store.New(c.db).CreateCalendarCode(
		ctx,
		store.CreateCalendarCodeParams{
			CalendarID: req.Msg.Id,
			Code:       req.Msg.Code,

			ExpiresAt: sql.NullTime{
				Time:  current,
				Valid: valid,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.CreateCodeResponse{}), nil
}

func (c *CalendarService) GetCodeMetadata(
	ctx context.Context,
	req *connect.Request[calendarsv1.GetCodeMetadataRequest],
) (*connect.Response[calendarsv1.GetCodeMetadataResponse], error) {
	var (
		row       store.GetCalendarCodeMetadataRow
		expiresAt *timestamppb.Timestamp
		err       error
	)

	row, err = store.New(c.db).GetCalendarCodeMetadata(ctx, req.Msg.CodeId)
	if err != nil {
		return nil, err
	}

	err = isCalendarOwner(ctx, c.db, row.CalendarID)
	if err != nil {
		return nil, err
	}

	if row.ExpiresAt.Valid {
		expiresAt = timestamppb.New(row.ExpiresAt.Time)
	}

	return connect.NewResponse(&calendarsv1.GetCodeMetadataResponse{
		CalendarId: row.CalendarID,
		Code:       row.Code,
		ExpiresAt:  expiresAt,
	}), nil
}

func (c *CalendarService) GetCodes(
	ctx context.Context,
	req *connect.Request[calendarsv1.GetCodesRequest],
) (*connect.Response[calendarsv1.GetCodesResponse], error) {
	var (
		codeIDs []int32
		err     error
	)

	err = isCalendarOwner(ctx, c.db, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	codeIDs, err = store.New(c.db).GetCalendarCodes(ctx, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.GetCodesResponse{
		CodeIds: codeIDs,
	}), nil
}

func (c *CalendarService) DeleteCode(
	ctx context.Context,
	req *connect.Request[calendarsv1.DeleteCodeRequest],
) (*connect.Response[calendarsv1.DeleteCodeResponse], error) {
	var (
		id  int32
		err error
	)

	id, err = store.New(c.db).GetCalendarCodeCalendarID(ctx, req.Msg.CodeId)
	if err != nil {
		return nil, err
	}

	err = isCalendarOwner(ctx, c.db, id)
	if err != nil {
		return nil, err
	}

	err = store.New(c.db).DeleteCalendarCode(ctx, req.Msg.CodeId)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.DeleteCodeResponse{}), nil
}

func (c *CalendarService) Subscribe(
	ctx context.Context,
	req *connect.Request[calendarsv1.SubscribeRequest],
) (*connect.Response[calendarsv1.SubscribeResponse], error) {
	var (
		actor   UserActor
		row     store.GetCalendarCodeFromCodeRow
		ownerID int32
		err     error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	row, err = store.New(c.db).GetCalendarCodeFromCode(ctx, req.Msg.Code)
	if err != nil {
		return nil, err
	}

	if row.ExpiresAt.Valid && row.DBTime.After(row.ExpiresAt.Time) {
		return nil, errors.New("code expired")
	}

	ownerID, err = store.New(c.db).GetCalendarOwner(ctx, row.CalendarID)
	if err != nil {
		return nil, err
	}

	if actor.ID == ownerID {
		return nil, errors.New("action not permitted")
	}

	err = store.New(c.db).SubscribeToCalendar(
		ctx,
		store.SubscribeToCalendarParams{
			UserID:     actor.ID,
			CalendarID: row.CalendarID,
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.SubscribeResponse{}), nil
}

func (c *CalendarService) Unsubscribe(
	ctx context.Context,
	req *connect.Request[calendarsv1.UnsubscribeRequest],
) (*connect.Response[calendarsv1.UnsubscribeResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = store.New(c.db).UnsubscribeFromCalendar(
		ctx,
		store.UnsubscribeFromCalendarParams{
			UserID:     actor.ID,
			CalendarID: req.Msg.Id,
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.UnsubscribeResponse{}), nil
}

func (c *CalendarService) GetMembers(
	ctx context.Context,
	req *connect.Request[calendarsv1.GetMembersRequest],
) (*connect.Response[calendarsv1.GetMembersResponse], error) {
	var (
		userIDs []int32
		err     error
	)

	err = isCalendarOwner(ctx, c.db, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	userIDs, err = store.New(c.db).GetCalendarMembers(ctx, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.GetMembersResponse{
		UserIds: userIDs,
	}), nil
}

func (c *CalendarService) RemoveMember(
	ctx context.Context,
	req *connect.Request[calendarsv1.RemoveMemberRequest],
) (*connect.Response[calendarsv1.RemoveMemberResponse], error) {
	var err error

	err = isCalendarOwner(ctx, c.db, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	err = store.New(c.db).UnsubscribeFromCalendar(
		ctx,
		store.UnsubscribeFromCalendarParams{
			UserID:     req.Msg.UserId,
			CalendarID: req.Msg.Id,
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.RemoveMemberResponse{}), nil
}

func (c *CalendarService) BanMember(
	ctx context.Context,
	req *connect.Request[calendarsv1.BanMemberRequest],
) (*connect.Response[calendarsv1.BanMemberResponse], error) {
	var (
		current time.Time
		valid   bool
		err     error
	)

	err = isCalendarOwner(ctx, c.db, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	current, err = store.New(c.db).GetCurrentTime(ctx)
	if err != nil {
		return nil, err
	}

	if req.Msg.Ttl != nil {
		current = current.Add(req.Msg.Ttl.AsDuration())
		valid = true
	}

	err = store.New(c.db).BanCalendarMember(
		ctx,
		store.BanCalendarMemberParams{
			UserID:     req.Msg.UserId,
			CalendarID: req.Msg.Id,
			Reason:     req.Msg.Reason,

			ExpiresAt: sql.NullTime{
				Valid: valid,
				Time:  current,
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.BanMemberResponse{}), nil
}

func (c *CalendarService) GetBannedMembers(
	ctx context.Context,
	req *connect.Request[calendarsv1.GetBannedMembersRequest],
) (*connect.Response[calendarsv1.GetBannedMembersResponse], error) {
	var (
		bannedUserIDs []int32
		err           error
	)

	err = isCalendarOwner(ctx, c.db, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	bannedUserIDs, err = store.New(c.db).GetCalendarBans(ctx, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.GetBannedMembersResponse{
		UserIds: bannedUserIDs,
	}), nil
}

func (c *CalendarService) UnbanMember(
	ctx context.Context,
	req *connect.Request[calendarsv1.UnbanMemberRequest],
) (*connect.Response[calendarsv1.UnbanMemberResponse], error) {
	var err error

	err = isCalendarOwner(ctx, c.db, req.Msg.Id)
	if err != nil {
		return nil, err
	}

	err = store.New(c.db).UnbanCalendarMember(
		ctx,
		store.UnbanCalendarMemberParams{
			UserID:     req.Msg.UserId,
			CalendarID: req.Msg.Id,
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&calendarsv1.UnbanMemberResponse{}), nil
}

func NewCalendarHandler(
	db *sql.DB,
	opts ...connect.HandlerOption,
) (string, http.Handler) {
	return calendarsv1connect.NewCalendarServiceHandler(
		NewCalendarService(db),
		opts...,
	)
}

func isValidColor(color string) error {
	if !reHex.MatchString(color) {
		return fmt.Errorf("invalid color hex: %s", color)
	}

	return nil
}

func isCalendarOwner(ctx context.Context, db store.DBTX, id int32) error {
	var (
		actor   UserActor
		ownerID int32
		err     error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return err
	}

	ownerID, err = store.New(db).GetCalendarOwner(ctx, id)
	if err != nil {
		return err
	}

	if actor.ID != ownerID {
		return errors.New("action not permitted")
	}

	return nil
}

func forEach[T any](supply func() []T, consume func(T)) {
	var elem T

	for _, elem = range supply() {
		consume(elem)
	}
}

func (c *CalendarService) mergeICals(
	oldICal []byte,
	newICal []byte,
) ([]byte, error) {
	var (
		oldCal, newCal *ics.Calendar
		err            error
	)

	oldCal, err = ics.ParseCalendar(bytes.NewReader(oldICal))
	if err != nil {
		return nil, err
	}

	newCal, err = ics.ParseCalendar(bytes.NewReader(newICal))
	if err != nil {
		return nil, err
	}

	forEach(newCal.Alarms, oldCal.AddVAlarm)
	forEach(newCal.Busys, oldCal.AddVBusy)
	forEach(newCal.Events, oldCal.AddVEvent)
	forEach(newCal.Journals, oldCal.AddVJournal)

	forEach(oldCal.Timezones, func(elem *ics.VTimezone) {
		var temp *ics.VTimezone

		temp = newCal.AddTimezone(elem.Id())
		*temp = *elem
	})

	forEach(oldCal.Todos, func(elem *ics.VTodo) {
		var temp *ics.VTodo

		temp = newCal.AddTodo(elem.Id())
		*temp = *elem
	})

	return []byte(newCal.Serialize()), nil
}
