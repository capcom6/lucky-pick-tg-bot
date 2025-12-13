package llm

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
