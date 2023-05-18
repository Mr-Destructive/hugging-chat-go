package main

import (
	"fmt"
	"os"

	"github.com/Mr-Destructive/hugging-chat-go/hugchat"
)

func main() {
	err := hugging_chat_go.LoadEnvFromFile(".env")
	cookies_map := map[string]string{"hf-chat": os.Getenv("hf-chat")}
	var inp string
	fmt.Println("Enter the prompt: ")
	fmt.Scanln(&inp)
	fmt.Println(inp)
	bot, err := hugging_chat_go.NewChatBot(cookies_map, "")
	if err != nil {
		fmt.Println(err)
	}

	text := inp
	temperature := 0.9
	topP := 0.95
	repetitionPenalty := 1.2
	topK := 50
	truncate := 1024
	watermark := false
	maxNewTokens := 1024
	stop := []string{"</s>"}
	returnFullText := false
	stream := true
	useCache := false
	isRetry := false
	retryCount := 5

	response, err := bot.Chat(text, temperature, topP, repetitionPenalty, topK, truncate, watermark, maxNewTokens, stop, returnFullText, stream, useCache, isRetry, retryCount)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Println(response)
}
