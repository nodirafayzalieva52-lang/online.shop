package ai

import (
	"context"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

const reviewSystemPromt = `Ты - генератор отзывов для интернет-магазиа. Твоя задача - создать реалистичныйб уникальный и естественный отзыв о товаре на основе параметров от пользователя.
Входные данные:
1. Название товара: {{PRODUCT_NAME}}
2. Тон отзыва: {{TONE}} (Варианты: Хвалебный / Ироничный и смешной / Скептический и критичный / В стихах / Честный и подробный)
3. Оценка (1-5 звёзд): {{RATING}}
Требования к отзыву:
1. Длина: от 2 до 5 предложений (или 2-3 четверостишия, если выбрали тон "В стихах").
2. Язык: Живой, естественный, с лёгкими разговорными оборотами (без канце ляризмов и шаблонных фраз вроде "данный товар превзошёл все мои ожидания").
3. Содержание: Упомяни 1-2 конкретные детали или сценария использования, характерные для этого товавра, чтобы отзыв выглядел убедительно.
4. Формат ответа: Верни ТОЛЬКО текст отзыва. Не добавляй никаких приветствий, заголовков, коментариев или кавычек.`

type ReviewTone string

const ( 
	Tonepraise  ReviewTone = "Хвалебный"
	ToneIronic  ReviewTone = "Ироничный и смешной"
	ToneSkeptic ReviewTone = "Скептическмй и критичный"
	TonePoetic  ReviewTone = "В стихах"
	ToneHonest  ReviewTone = "Честный и подробный"
)

func GenerateReview(productName string, tone ReviewTone, rating int) (string, error) {
	if productName == "" {
		return "", fmt.Errorf("product name is required")
	}
	if rating < 1 || rating > 5 {
		return "", fmt.Errorf("rating must be between 1 and 5")
	}

	client := openai.NewClient("YOUR_OPENAI_API_KEY")
	ctx := context.Background()

	userPromt := fmt.Sprintf(
		"Название товара: %s\nТон отзыва: %s\nОценка: %d",
		productName, tone, rating,
	)

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: openai.GPT4oMini,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: reviewSystemPromt},
			{Role: openai.ChatMessageRoleUser, Content: userPromt},
		},
		Temperature: 0.9,
	})
	if err != nil {
		return "", err
	}

	return resp.Choices[0].Message.Content, nil
}