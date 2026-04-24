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
	"github.com/andrieee44/countmein/gen/ai/v2"
	"github.com/andrieee44/countmein/gen/ai/v2/aiv2connect"
	ics "github.com/arran4/golang-ical"
	groq "github.com/conneroisu/groq-go"
	"github.com/google/uuid"
)

const (
	temperature = 0.1
	maxTokens   = 1024
)

var (
	reUID *regexp.Regexp = regexp.MustCompile(`UID:\{\?\}`)
	model groq.ChatModel = groq.ModelLlama3370BVersatile

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

func (a *AIService) TextToICal(
	ctx context.Context,
	req *connect.Request[aiv2.TextToICalRequest],
) (*connect.Response[aiv2.TextToICalResponse], error) {
	var (
		client *groq.Client
		text   string
		ical   *ics.Calendar
		resp   groq.ChatCompletionResponse
		err    error
	)

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
		return nil, err
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

	ical, err = ics.ParseCalendar(strings.NewReader(text))
	if err != nil {
		return nil, err
	}

	return connect.NewResponse(&aiv2.TextToICalResponse{
		Ical: []byte(ical.Serialize()),
	}), nil
}

func (a *AIService) AnalyzeICal(
	ctx context.Context,
	req *connect.Request[aiv2.AnalyzeICalRequest],
) (*connect.Response[aiv2.AnalyzeICalResponse], error) {
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

	return connect.NewResponse(&aiv2.AnalyzeICalResponse{
		Analysis: resp.Choices[0].Message.Content,
	}), nil
}

func AIHandler(
	logger *slog.Logger,
	opts ...connect.HandlerOption,
) RPCHandlerFn {
	return func() (string, http.Handler) {
		return aiv2connect.NewAIServiceHandler(
			NewAIService(logger),
			opts...,
		)
	}
}
