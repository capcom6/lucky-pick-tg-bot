package giveaways

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/revrost/go-openrouter"
	"github.com/revrost/go-openrouter/jsonschema"
	"go.uber.org/zap"
)

const (
	DescriptionGenerationPrompt = `Ты - администратор Telegram-группы розыгрышей.
Твоя задача: придумывать веселое описание товара на фотографии длиной не более 150 слов.

Описание от пользователя: {description}
Отвечай от имени персонажа, связанного с праздником, ближайшим к дате: {publish_date}

Требования:
- Ответ должен быть на русском языке
- Используй только широко известные праздники в России
- Избегай религиозных праздников
- Не упоминай, что ты бот
- Не упоминай никакие даты`
)

type LLMDescriptionAnswer struct {
	Description string `json:"description" description:"Описание товара" required:"true"` //
}

type LLM struct {
	config Config

	client *openrouter.Client

	logger *zap.Logger
}

func NewLLM(config Config, client *openrouter.Client, logger *zap.Logger) *LLM {
	return &LLM{
		config: config,

		client: client,

		logger: logger,
	}
}

func (l *LLM) MakeDescription(
	ctx context.Context,
	description string,
	publishDate time.Time,
	image []byte,
) (string, error) {
	l.logger.Debug("making description",
		zap.String("description", description),
		zap.Time("publish_date", publishDate),
	)

	replacer := strings.NewReplacer(
		"{description}", description,
		"{publish_date}", publishDate.Format(time.DateOnly),
	)
	prompt := replacer.Replace(DescriptionGenerationPrompt)

	answer := new(LLMDescriptionAnswer)
	schema, err := jsonschema.GenerateSchemaForType(answer)
	if err != nil {
		return "", fmt.Errorf("failed to generate schema: %w", err)
	}

	request := openrouter.ChatCompletionRequest{
		Model: l.config.LLMModel,
		Messages: []openrouter.ChatCompletionMessage{
			{
				Role: openrouter.ChatMessageRoleUser,
				Content: openrouter.Content{
					Multi: []openrouter.ChatMessagePart{
						{
							Type: openrouter.ChatMessagePartTypeText,
							Text: prompt,
						},
						{
							Type: openrouter.ChatMessagePartTypeImageURL,
							ImageURL: &openrouter.ChatMessageImageURL{
								URL: "data:" + http.DetectContentType(
									image,
								) + ";base64," + base64.StdEncoding.EncodeToString(
									image,
								),
							},
						},
					},
				},
			},
		},
		ResponseFormat: &openrouter.ChatCompletionResponseFormat{
			Type: openrouter.ChatCompletionResponseFormatTypeJSONSchema,
			JSONSchema: &openrouter.ChatCompletionResponseFormatJSONSchema{
				Name:   "description",
				Schema: schema,
				Strict: true,
			},
		},
	}

	res, err := l.client.CreateChatCompletion(ctx, request)
	if err != nil {
		return "", fmt.Errorf("failed to generate description: %w", err)
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("%w: no choices returned", ErrLLMFailed)
	}

	if jsonErr := json.Unmarshal([]byte(res.Choices[0].Message.Content.Text), answer); jsonErr != nil {
		return "", fmt.Errorf("failed to unmarshal answer: %w", jsonErr)
	}

	return answer.Description, nil
}
