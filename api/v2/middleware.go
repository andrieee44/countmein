package api

import (
	"context"
	"log/slog"

	"connectrpc.com/connect"
)

func ErrorInterceptor(logger *slog.Logger) connect.UnaryInterceptorFunc {
	return func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(
			ctx context.Context,
			req connect.AnyRequest,
		) (connect.AnyResponse, error) {
			var (
				res connect.AnyResponse
				err error
			)

			res, err = next(ctx, req)
			if err == nil {
				return res, nil
			}

			logger.Error(
				"RPC Request Failed",
				"procedure", req.Spec().Procedure,
				"protocol", req.Peer().Protocol,
				"error", err,
			)

			return nil, err
		}
	}
}
