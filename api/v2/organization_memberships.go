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

type OrganizationMembershipService struct {
	db *sql.DB
}

func NewOrganizationMembershipService(
	db *sql.DB,
) *OrganizationMembershipService {
	return &OrganizationMembershipService{
		db: db,
	}
}

func (o *OrganizationMembershipService) JoinOrganization(
	ctx context.Context,
	req *connect.Request[organizationsv2.JoinOrganizationRequest],
) (*connect.Response[organizationsv2.JoinOrganizationResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = store.New(o.db).JoinOrganization(
		ctx,
		store.JoinOrganizationParams{
			ActorUserID:    actor.UserID,
			OrganizationID: req.Msg.OrganizationId,
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(
		&organizationsv2.JoinOrganizationResponse{},
	), nil
}

func (o *OrganizationMembershipService) LeaveOrganization(
	ctx context.Context,
	req *connect.Request[organizationsv2.LeaveOrganizationRequest],
) (*connect.Response[organizationsv2.LeaveOrganizationResponse], error) {
	var (
		actor UserActor
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = store.New(o.db).LeaveOrganization(
		ctx,
		store.LeaveOrganizationParams{
			ActorUserID:    actor.UserID,
			OrganizationID: req.Msg.OrganizationId,
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(
		&organizationsv2.LeaveOrganizationResponse{},
	), nil
}

func OrganizationMembershipHandler(
	db *sql.DB,
	opts ...connect.HandlerOption,
) RPCHandlerFn {
	return func() (string, http.Handler) {
		return organizationsv2connect.NewOrganizationMembershipServiceHandler(
			NewOrganizationMembershipService(db),
			opts...,
		)
	}
}
