package hugchat

import (
	"net/http"
	"net/url"
)

type ChatBot struct {
	Username             string
	Password             string
	Cookies              []*http.Cookie
	Session              *http.Client
	HFBaseURL            *url.URL
	JSONHeader           http.Header
	ConversationIDList   []string
	ActiveModel          string
	AcceptedWelcomeModal bool
	CurrentConversation  string
}

type ChatParameters struct {
	Temperature       float64  `json:"temperature"`
	TopP              float64  `json:"top_p"`
	RepetitionPenalty float64  `json:"repetition_penalty"`
	TopK              int      `json:"top_k"`
	Truncate          int      `json:"truncate"`
	Watermark         bool     `json:"watermark"`
	MaxNewTokens      int      `json:"max_new_tokens"`
	Stop              []string `json:"stop"`
	ReturnFullText    bool     `json:"return_full_text"`
	Stream            bool     `json:"stream"`
}

type ChatOptions struct {
	UseCache bool   `json:"use_cache"`
	IsRetry  bool   `json:"is_retry"`
	ID       string `json:"id"`
}

type ChatRequest struct {
	Inputs     string         `json:"inputs"`
	Parameters ChatParameters `json:"parameters"`
	Options    ChatOptions    `json:"options"`
}

type ChatResponse struct {
	GeneratedText string `json:"generated_text"`
	Error         string `json:"error"`
}
