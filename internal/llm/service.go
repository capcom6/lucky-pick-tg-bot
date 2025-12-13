package llm

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"time"

	"github.com/revrost/go-openrouter"
	"github.com/revrost/go-openrouter/jsonschema"
	"go.uber.org/zap"
)

// Service provides LLM-powered functionality for generating moderation questions
// and managing interactions with the OpenRouter API.
type Service struct {
	config Config

	client *openrouter.Client

	logger *zap.Logger
}

// NewService creates a new LLM Service instance with the provided OpenRouter client and logger.
func NewService(config Config, client *openrouter.Client, logger *zap.Logger) *Service {
	return &Service{
		config: config,

		client: client,

		logger: logger,
	}
}

func (s *Service) MakeQuestion(ctx context.Context, description string, age time.Duration) (string, error) {
	s.logger.Debug("making question",
		zap.String("description", description),
		zap.Duration("age", age),
	)

	replacer := strings.NewReplacer(
		"{lot_description}", description,
		"{hours_since_post}", strconv.Itoa(int(age.Hours())),
		"{current_date}", time.Now().Format(time.DateOnly),
	)
	prompt := replacer.Replace(GiveawayQuestionPrompt)

	answer := new(GiveawayQuestionAnswer)
	schema, err := jsonschema.GenerateSchemaForType(answer)
	if err != nil {
		return "", fmt.Errorf("failed to generate schema: %w", err)
	}

	request := openrouter.ChatCompletionRequest{
		Model: s.config.Model,
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role: openrouter.ChatMessageRoleUser,
				Content: openrouter.Content{
					Text: prompt,
				},
			},
		},
		ResponseFormat: &openrouter.ChatCompletionResponseFormat{
			Type: openrouter.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openrouter.ChatCompletionResponseFormatJSONSchema{
				Name:   "questions",
				Schema: schema,
				Strict: true,
			},
		},
	}

	res, err := s.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to create chat completion: %w", err)
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("%w: no choices returned", ErrLLMFailed)
	}

	if jsonErr := json.Unmarshal([]byte(res.Choices[0].Message.Content.Text), answer); jsonErr != nil {
		return "", fmt.Errorf("failed to unmarshal answer: %w", jsonErr)
	}

	const questionsCount = 3
	i, err := rand.Int(rand.Reader, big.NewInt(questionsCount))
	if err != nil {
		return "", fmt.Errorf("failed to generate random number: %w", err)
	}

	switch i.Int64() {
	case 0:
		return answer.OpenQuestion, nil
	case 1:
		return answer.QuickAnswer, nil
	default:
		return answer.CreativeTask, nil
	}
}
