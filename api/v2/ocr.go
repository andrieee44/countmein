package api

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/andrieee44/countmein/gen/ocr/v2"
	"github.com/andrieee44/countmein/gen/ocr/v2/ocrv2connect"
	"github.com/otiai10/gosseract/v2"
)

type OCRService struct{}

func NewOCRService() *OCRService {
	return &OCRService{}
}

func (o *OCRService) ImageToText(
	ctx context.Context,
	req *connect.Request[ocrv2.ImageToTextRequest],
) (*connect.Response[ocrv2.ImageToTextResponse], error) {
	var (
		client *gosseract.Client
		text   string
		err    error
	)

	client = gosseract.NewClient()
	defer client.Close()

	err = client.SetImageFromBytes(req.Msg.Image)
	if err != nil {
		return nil, err
	}

	text, err = client.Text()
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&ocrv2.ImageToTextResponse{
		Text: text,
	}), nil
}

func OCRHandler(opts ...connect.HandlerOption) RPCHandlerFn {
	return func() (string, http.Handler) {
		return ocrv2connect.NewOCRServiceHandler(
			NewOCRService(),
			opts...,
		)
	}
}
