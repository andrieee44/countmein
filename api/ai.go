package api

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"connectrpc.com/connect"
	aiv1 "github.com/andrieee44/countmein/gen/ai/v1"
	"github.com/andrieee44/countmein/gen/ai/v1/aiv1connect"
	ics "github.com/arran4/golang-ical"
	groq "github.com/conneroisu/groq-go"
	"github.com/google/uuid"
	"github.com/otiai10/gosseract/v2"
)

const (
	temperature = 0.1
	maxTokens   = 1024
)

var (
	reSpaces *regexp.Regexp = regexp.MustCompile(`[ \t]+`)
	reLines  *regexp.Regexp = regexp.MustCompile(`\n{3,}`)
	reUID    *regexp.Regexp = regexp.MustCompile(`UID:\{\?\}`)
	model    groq.ChatModel = groq.ModelLlama3370BVersatile

	//go:embed prompts/text_to_ical.md
	textToICalSystemPrompt string

	//go:embed prompts/ical_to_analysis.md
	icalToAnalysisSystemPrompt string
)

type AIService struct {
	logger *slog.Logger
}

func NewAIService(logger *slog.Logger) *AIService {
	return &AIService{
		logger: logger,
	}
}

func (a *AIService) ImageToICal(
	ctx context.Context,
	req *connect.Request[aiv1.ImageToICalRequest],
) (*connect.Response[aiv1.ImageToICalResponse], error) {
	var (
		text, icalText string
		ical           *ics.Calendar
		err            error
	)

	text, err = a.imageToText(req.Msg.Image)
	if err != nil {
		return nil, err
	}

	icalText, err = a.textToICalText(ctx, text)
	if err != nil {
		return nil, err
	}

	ical, err = ics.ParseCalendar(strings.NewReader(icalText))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&aiv1.ImageToICalResponse{
		Ical: []byte(ical.Serialize()),
	}), nil
}

func (a *AIService) AnalyzeICal(
	ctx context.Context,
	req *connect.Request[aiv1.AnalyzeICalRequest],
) (*connect.Response[aiv1.AnalyzeICalResponse], error) {
	var (
		client *groq.Client
		resp   groq.ChatCompletionResponse
		err    error
	)

	_, err = ics.ParseCalendar(bytes.NewReader(req.Msg.Ical))
	if err != nil {
		return nil, err
	}

	client, err = groq.NewClient(
		os.Getenv("GROQ_API_KEY"),
		groq.WithLogger(a.logger),
	)
	if err != nil {
		return nil, err
	}

	resp, err = client.ChatCompletion(
		ctx,
		groq.ChatCompletionRequest{
			Model:       model,
			Temperature: temperature,
			MaxTokens:   maxTokens,

			Messages: []groq.ChatCompletionMessage{
				{
					Role:    groq.RoleSystem,
					Content: icalToAnalysisSystemPrompt,
				},
				{
					Role:    groq.RoleUser,
					Content: req.Msg.String(),
				},
			},
		},
	)
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&aiv1.AnalyzeICalResponse{
		Analysis: resp.Choices[0].Message.Content,
	}), nil
}

func (a *AIService) imageToText(buf []byte) (string, error) {
	var (
		client     *gosseract.Client
		mime, text string
		err        error
	)

	mime = http.DetectContentType(buf)

	switch mime {
	case "image/png",
		"image/jpeg",
		"image/gif",
		"image/webp",
		"image/tiff",
		"image/bmp":
	default:
		return "", fmt.Errorf("unsupported image type: %s", mime)
	}

	client = gosseract.NewClient()
	defer client.Close()

	err = client.SetImageFromBytes(buf)
	if err != nil {
		return "", err
	}

	text, err = client.Text()
	if err != nil {
		return "", err
	}

	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = reSpaces.ReplaceAllString(text, " ")
	text = reLines.ReplaceAllString(text, "\n\n")

	return strings.TrimSpace(text), nil
}

func (a *AIService) textToICalText(
	ctx context.Context,
	text string,
) (string, error) {
	var (
		client *groq.Client
		resp   groq.ChatCompletionResponse
		err    error
	)

	client, err = groq.NewClient(
		os.Getenv("GROQ_API_KEY"),
		groq.WithLogger(a.logger),
	)
	if err != nil {
		return "", err
	}

	resp, err = client.ChatCompletion(
		ctx,
		groq.ChatCompletionRequest{
			Model:       model,
			Temperature: temperature,
			MaxTokens:   maxTokens,

			Messages: []groq.ChatCompletionMessage{
				{
					Role:    groq.RoleSystem,
					Content: textToICalSystemPrompt,
				},
				{
					Role: groq.RoleAssistant,

					Content: fmt.Sprintf(
						"Current time (DTSTAMP): %s",
						time.Now().UTC().Format("20060102T150405Z"),
					),
				},
				{
					Role:    groq.RoleUser,
					Content: text,
				},
			},
		},
	)
	if err != nil {
		return "", err
	}

	text = strings.ReplaceAll(
		resp.Choices[0].Message.Content,
		"PRODID:{?}",
		"PRODID:-//andrieee44//countmein//EN",
	)

	text = reUID.ReplaceAllStringFunc(text, func(s string) string {
		return fmt.Sprintf(
			"UID:%s@github.com/andrieee44/countmein",
			uuid.New(),
		)
	})

	return text, nil
}

func NewAIHandler(
	logger *slog.Logger,
	opts ...connect.HandlerOption,
) (string, http.Handler) {
	return aiv1connect.NewAIServiceHandler(
		NewAIService(logger),
		opts...,
	)
}
