package discussions

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

const (
	// GiveawayQuestionPrompt is a Russian-language prompt template for generating moderation questions
	// about giveaway lots in a Telegram group. It contains placeholders:
	//   - {lot_description}: description of the current lot
	//   - {hours_since_post}: hours elapsed since the lot was posted
	//   - {current_date}: current date
	GiveawayQuestionPrompt = `Ты - модератор Telegram-группы розыгрышей.
Твоя задача: поддерживать легкую активность, задавая вопросы о лотах.

Описание последнего лота: {lot_description}
Время с публикации: {hours_since_post} часов
Отвечай в стиле праздника, ближайшего к текущей дате: {current_date}

Сгенерируй:
1. Один открытый вопрос для обсуждения
2. Один легкий вопрос для быстрого ответа
3. Один творческий вопрос/задание

Требования:
- Ответ должен быть на русском языке
- Не отвечай на вопросы сам
- Не упоминай, что ты бот
- Вопросы должны быть естественными
- Избегай спама и навязчивости`
)

type GiveawayQuestionAnswer struct {
	OpenQuestion string `json:"open_question" description:"Открытый вопрос для обсуждения"    required:"true"` // Открытый вопрос для обсуждения
	QuickAnswer  string `json:"quick_answer"  description:"Легкий вопрос для быстрого ответа" required:"true"` // Легкий вопрос для быстрого ответа
	CreativeTask string `json:"creative_task" description:"Творческий вопрос/задание"         required:"true"` // Творческий вопрос/задание
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

func (s *LLM) MakeQuestion(ctx context.Context, description string, age time.Duration) (string, error) {
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
		Model: s.config.LLMModel,
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
