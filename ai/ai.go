package ai

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type SearchArgs struct {
	Query    string  `json:"query"`
	MaxPrice float64 `json:"max_price,omitempty"`
}

type OrderArgs struct {
	OrderID string `json:"order_id"`
}

func searchProductsDB(query string, maxPrice float64) string {
	return `[{"id": 101, "name": "Черная худи Go", "price": 150.00, "in_stock": true}]`
}

func getOrderStatusDB(orderID string) string {
	return fmt.Sprintf(`{"order_id": "%s", "status": "В пути", "delivery_date": "Завтра"}`, orderID)
}

func AskAI(userMessage string) (string, error) {
	client := openai.NewClient("YOUR_OPENAI_API_KEY")
	ctx := context.Background()

	tools := []openai.Tool{
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "search_products",
				Description: "Поиск товаров в каталоге",
				Parameters: json.RawMessage([]byte(`{
					"type": "object",
					"properties": {
						"query": {"type": "string"},
						"max_price": {"type": "number"}
					},
					"required": ["query"]
				}`)),
			},
		},
		{
			Type: openai.ToolTypeFunction,
			Function: &openai.FunctionDefinition{
				Name:        "get_order_status",
				Description: "Получение статуса заказа",
				Parameters: json.RawMessage([]byte(`{
					"type": "object",
					"properties": {
						"order_id": {"type": "string"}
					},
					"required": ["order_id"]
				}`)),
			},
		},
	}

	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: "Ты — консультант интернет-магазина. Для поиска товаров и заказов вызывай функции.",
		},
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userMessage,
		},
	}

	resp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model:    openai.GPT4oMini,
		Messages: messages,
		Tools:    tools,
	})
	if err != nil {
		return "", err
	}

	msg := resp.Choices[0].Message

	if len(msg.ToolCalls) > 0 {
		toolCall := msg.ToolCalls[0]
		messages = append(messages, msg)

		var toolResult string
		if toolCall.Function.Name == "search_products" {
			var args SearchArgs
			_ = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
			toolResult = searchProductsDB(args.Query, args.MaxPrice)
		} else if toolCall.Function.Name == "get_order_status" {
			var args OrderArgs
			_ = json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
			toolResult = getOrderStatusDB(args.OrderID)
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:       openai.ChatMessageRoleTool,
			Content:    toolResult,
			ToolCallID: toolCall.ID,
		})

		finalResp, err := client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model:    openai.GPT4oMini,
			Messages: messages,
		})
		if err != nil {
			return "", err
		}

		return finalResp.Choices[0].Message.Content, nil
	}

	return msg.Content, nil
}