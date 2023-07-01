# Hugging Face Chatbot Example

This is a basic example of using the HuggingFace Chat Client in Go.

## Installation

```
go get github.com/mr-destructive/hugging-chat-go
```

##  Usage

Add a `.env` file with the credentials of the hugchat application.

```
email=abc@def.com
password=superSecret
```

Use the `hugchat.NewChatBot(email, password)` method to instantiate the chat bot object by authenticating the sesssion. Further with that object, request the api with the `Chat` method by fine tuning the parameters provided in the below example:


```go
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/mr-destructive/hugging-chat-go/hugchat"
)

func main() {

    err := hugchat.LoadEnvFromFile(".env")
	//cookies_map := map[string]string{"hf-chat": os.Getenv("hf-chat")}
	email := os.Getenv("email")
	password := os.Getenv("password")
	var inp string
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Enter the prompt:")
	inp, _ = reader.ReadString('\n')
	bot, err := hugchat.NewChatBot(email, password)
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
```

