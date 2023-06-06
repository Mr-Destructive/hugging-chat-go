package hugchat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

func NewChatBot(username, password string) (*ChatBot, error) {
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}

	baseURL, err := url.Parse("https://huggingface.co")
	if err != nil {
		return nil, fmt.Errorf("failed to parse base URL: %w", err)
	}

	cb := ChatBot{
		Session:              &http.Client{Timeout: time.Second * 10},
		HFBaseURL:            baseURL,
		JSONHeader:           http.Header{"Content-Type": []string{"application/json"}},
		ConversationIDList:   []string{},
		ActiveModel:          "OpenAssistant/oasst-sft-6-llama-30b-xor",
		AcceptedWelcomeModal: false,
		CurrentConversation:  "",
	}

	login := NewLogin(username, password, "usercookies.json")
	if err := login.Login(); err != nil {
		fmt.Println("Login failed:", err)
		return nil, err
	}

	if err := login.SaveCookies(); err != nil {
		fmt.Println("Failed to save cookies:", err)
		return nil, err
	}
	cb.Cookies = login.Cookies
	cb.AcceptEthicsModal()
	err = cb.setHCSession()
	if err != nil {
		return nil, fmt.Errorf("failed to set HC session: %w", err)
	}

	cb.CurrentConversation, err = cb.NewConversation()
	if err != nil {
		return nil, fmt.Errorf("failed to create new conversation: %w", err)
	}

	return &cb, nil
}

func (cb *ChatBot) setHCSession() error {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return fmt.Errorf("failed to create cookie jar: %w", err)
	}

	//cb.Session.Jar = jar
	cb.Session = &http.Client{Timeout: time.Second * 10, Jar: jar}
	cb.Session.Jar.SetCookies(cb.HFBaseURL, cb.Cookies)
	cb.Session.Get(cb.HFBaseURL.String() + "/")

	//baseURL := cb.HFBaseURL.String() + "/"
	//cb.Session.Get(baseURL)

	return nil
}

func makeCookies(cookies map[string]string) []*http.Cookie {
	cookieList := make([]*http.Cookie, 0, len(cookies))
	for name, value := range cookies {
		cookieList = append(cookieList, &http.Cookie{
			Name:  name,
			Value: value,
		})
	}
	return cookieList
}

func (c *ChatBot) NewConversation() (string, error) {
	errCount := 0

	url := fmt.Sprintf("%s/chat/conversation", c.HFBaseURL.String())
	data := fmt.Sprintf(`{"model": "%s"}`, c.ActiveModel)
	req, err := http.NewRequest("POST", url, strings.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.Session.Do(req)
	if err != nil {
		errCount++
		if errCount > 5 {
			return "", fmt.Errorf("failed to create new conversation: %w", err)
		}
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		errCount++
		if errCount > 5 {
			return "", fmt.Errorf("failed to create new conversation with status code %d", resp.StatusCode)
		}
	}

	var response struct {
		ConversationID string `json:"conversationId"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal response body: %w", err)
	}

	c.ConversationIDList = append(c.ConversationIDList, response.ConversationID)
	return response.ConversationID, nil
}

func (c *ChatBot) AcceptEthicsModal() error {
	url := c.HFBaseURL.String() + "/chat/settings"
	data := strings.NewReader(fmt.Sprintf(`{"ethicsModalAccepted": "true","shareConversationsWithModelAuthors": "true","ethicsModalAcceptedAt": "","activeModel": "%s"}`, c.ActiveModel))
	req, err := http.NewRequest("POST", url, data)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	if !c.AcceptedWelcomeModal {
		return errors.New("Welcome modal not accepted")
	}

	resp, err := c.Session.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("Failed to accept ethics modal with status code %d. %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *ChatBot) getHeaders(ref bool) http.Header {
	headers := make(http.Header)
	headers.Set("Host", "api.huggingface.com")
	headers.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/96.0.4664.93 Safari/537.36")
	headers.Set("Accept", "application/json")
	headers.Set("Accept-Language", "en-US,en;q=0.9")
	headers.Set("Referer", fmt.Sprintf("%s/chat", c.HFBaseURL.String()))

	if ref {
		headers.Set("Referer", fmt.Sprintf("%s/chat/conversation/%s", c.HFBaseURL.String(), c.CurrentConversation))
	}

	return headers
}

func (c *ChatBot) Chat(
	text string,
	temperature float64,
	topP float64,
	repetitionPenalty float64,
	topK int,
	truncate int,
	watermark bool,
	maxNewTokens int,
	stop []string,
	returnFullText bool,
	stream bool,
	useCache bool,
	isRetry bool,
	retryCount int,
) (string, error) {
	if retryCount <= 0 {
		return "", errors.New("retryCount must be greater than 0")
	}

	conversationID, err := GenerateUUID()
	if err != nil {
		return "", fmt.Errorf("failed to generate conversation ID: %w", err)
	}

	req := ChatRequest{
		Inputs: text,
		Parameters: ChatParameters{
			Temperature:       temperature,
			TopP:              topP,
			RepetitionPenalty: repetitionPenalty,
			TopK:              topK,
			Truncate:          truncate,
			Watermark:         watermark,
			MaxNewTokens:      maxNewTokens,
			Stop:              stop,
			ReturnFullText:    returnFullText,
			Stream:            stream,
		},
		Options: ChatOptions{
			UseCache: useCache,
			IsRetry:  isRetry,
			ID:       conversationID,
		},
	}

	url := fmt.Sprintf("%s/chat/conversation/%s", c.HFBaseURL.String(), c.CurrentConversation)
	reqData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request data: %w", err)
	}

	for retryCount > 0 {
		req, err := http.NewRequest("POST", url, bytes.NewBuffer(reqData))
		if err != nil {
			return "", fmt.Errorf("failed to create HTTP request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.Session.Do(req)
		if err != nil {
			retryCount--
			if retryCount <= 0 {
				return "", fmt.Errorf("failed to send HTTP request: %w", err)
			}
			continue
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", fmt.Errorf("failed to read response body: %w", err)
		}

		if resp.StatusCode != http.StatusOK {
			retryCount--
			if retryCount <= 0 {
				return "", fmt.Errorf("chat request failed with status code %d: %s", resp.StatusCode, body)
			}
			continue
		}

		var response []ChatResponse
		err = json.Unmarshal(body, &response)
		if err != nil {
			return "", fmt.Errorf("failed to unmarshal response body: %w", err)
		}

		if len(response) > 0 {
			for _, chatResp := range response {
				if chatResp.Error != "" {
					return "", fmt.Errorf("chat error: %s", chatResp.Error)
				}
			}
			return response[0].GeneratedText, nil
		}
	}

	return "", errors.New("chat request failed")
}
