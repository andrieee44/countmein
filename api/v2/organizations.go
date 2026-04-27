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
	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrganizationService struct {
	db *sql.DB
}

func NewOrganizationService(db *sql.DB) *OrganizationService {
	return &OrganizationService{
		db: db,
	}
}

func (o *OrganizationService) CreateOrganization(
	ctx context.Context,
	req *connect.Request[organizationsv2.CreateOrganizationRequest],
) (*connect.Response[organizationsv2.CreateOrganizationResponse], error) {
	var (
		actor          UserActor
		organizationID int64
		err            error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = sproc(
		ctx,
		o.db,
		"@out_organization_id",
		&organizationID,
		func(tx *sql.Tx) error {
			return store.New(tx).CreateOrganization(
				ctx,
				store.CreateOrganizationParams{
					ActorUserID:         actor.UserID,
					Name:                req.Msg.Name,
					RequiresJoinRequest: req.Msg.RequiresJoinRequest,
					Description:         req.Msg.Description,
				},
			)
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&organizationsv2.CreateOrganizationResponse{
		OrganizationId: organizationID,
	}), nil
}

func (o *OrganizationService) GetOrganization(
	ctx context.Context,
	req *connect.Request[organizationsv2.GetOrganizationRequest],
) (*connect.Response[organizationsv2.GetOrganizationResponse], error) {
	var (
		row store.GetOrganizationRow
		err error
	)

	row, err = store.New(o.db).GetOrganization(
		ctx,
		req.Msg.OrganizationId,
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&organizationsv2.GetOrganizationResponse{
		Name:                row.Name,
		RequiresJoinRequest: row.RequiresJoinRequest,
		CreatedAt:           timestamppb.New(row.CreatedAt),
		Description:         toPtr(row.Description),
	}), nil
}

func (o *OrganizationService) GetOrganizations(
	ctx context.Context,
	req *connect.Request[organizationsv2.GetOrganizationsRequest],
) (*connect.Response[organizationsv2.GetOrganizationsResponse], error) {
	var (
		organizationIDs []int64
		err             error
	)

	organizationIDs, err = store.New(o.db).GetOrganizations(ctx)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&organizationsv2.GetOrganizationsResponse{
		OrganizationIds: organizationIDs,
	}), nil
}

func (o *OrganizationService) UpdateOrganization(
	ctx context.Context,
	req *connect.Request[organizationsv2.UpdateOrganizationRequest],
) (*connect.Response[organizationsv2.UpdateOrganizationResponse], error) {
	var (
		actor UserActor
		n     int64
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	n, err = store.New(o.db).UpdateOrganization(
		ctx,
		store.UpdateOrganizationParams{
			ActorUserID:         actor.UserID,
			OrganizationID:      req.Msg.OrganizationId,
			Name:                fromPtr(req.Msg.Name),
			Description:         fromPtr(req.Msg.Description),
			RequiresJoinRequest: fromPtr(req.Msg.RequiresJoinRequest),
		},
	)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, ErrNotFoundOrDenied
	}

	return connect.NewResponse(
		&organizationsv2.UpdateOrganizationResponse{},
	), nil
}

func (o *OrganizationService) DeleteOrganization(
	ctx context.Context,
	req *connect.Request[organizationsv2.DeleteOrganizationRequest],
) (*connect.Response[organizationsv2.DeleteOrganizationResponse], error) {
	var (
		actor UserActor
		n     int64
		err   error
	)

	actor, err = ActorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	n, err = store.New(o.db).DeleteOrganization(
		ctx,
		store.DeleteOrganizationParams{
			ActorUserID:         actor.UserID,
			OrganizationID:      req.Msg.OrganizationId,
		},
	)
	if err != nil {
		return nil, err
	}

	if n == 0 {
		return nil, ErrNotFoundOrDenied
	}

	return connect.NewResponse(
		&organizationsv2.DeleteOrganizationResponse{},
	), nil
}

func OrganizationHandler(
	db *sql.DB,
	opts ...connect.HandlerOption,
) RPCHandlerFn {
	return func() (string, http.Handler) {
		return organizationsv2connect.NewOrganizationServiceHandler(
			NewOrganizationService(db),
			opts...,
		)
	}
}
