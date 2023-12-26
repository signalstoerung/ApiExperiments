package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

const (
	chatCompletionEndpoint = "https://api.openai.com/v1/chat/completions"
	ModelGPT3Latest        = "gpt-3.5-turbo-1106"
	ModelGPT3Standard      = "gpt-3.5-turbo"
	ModelGPT4Latest        = "gpt-4-1106-preview"
	ModelGPT4Standard      = "gpt-4"
)

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// dollars per 1000 tokens
var pricing = map[string]float64{
	"gpt-3.5-turbo":      0.0020,
	"gpt-3.5-turbo-1106": 0.0020,
	"gpt-4":              0.06,
	"gpt-4-1106-preview": 0.03,
}

var ApiKey string = ""

type Message struct {
	Name    string `json:"name,omitempty"`
	Content string `json:"content"`
	Role    Role   `json:"role"`
}

type Request struct {
	Messages  []Message `json:"messages"` // required
	Model     string    `json:"model"`    // required
	MaxTokens int       `json:"max_tokens,omitempty"`
}

type CompletionChoice struct {
	Index        int     `json:"index"`
	Message      Message `json:"message"`
	FinishReason string  `json:"finish_reason"`
}

type Usage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type Completion struct {
	Id      string             `json:"id"`
	Choices []CompletionChoice `json:"choices"`
	Created int                `json:"created"`
	Model   string             `json:"model"`
	Usage   Usage              `json:"usage"`
}

// calculate cost for completion
func (c Completion) Cost() (float64, error) {
	price, ok := pricing[c.Model]
	if !ok {
		return 0, fmt.Errorf("no price found for model %s", c.Model)
	}
	return price * float64(c.Usage.TotalTokens) / 1000, nil
}

func chatCompletion(request Request) (Completion, error) {
	completion := Completion{}
	reqBody, err := json.Marshal(request)
	//log.Printf("Json request object: %+v", string(reqBody))
	if err != nil {
		return completion, err
	}
	client := http.Client{}
	r, err := http.NewRequest(http.MethodPost, chatCompletionEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return completion, err
	}

	r.Header.Set("Authorization", "Bearer "+ApiKey)
	r.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(r)
	if err != nil {
		return completion, err
	}

	err = json.NewDecoder(resp.Body).Decode(&completion)
	if err != nil {
		return completion, err
	}

	if resp.StatusCode != http.StatusOK {
		return completion, fmt.Errorf("API request failed with status code %d: %s", resp.StatusCode, resp.Status)
	}
	cost, err := completion.Cost()
	if err != nil {
		log.Printf("Completion generated, %v tokens (%v)", completion.Usage.TotalTokens, err)
	} else {
		log.Printf("Completion generated at cost of $%.4f", cost)
	}
	return completion, nil
}

func HeadlineForText(text string) (string, error) {

	request := Request{
		Model:     ModelGPT3Latest,
		MaxTokens: 500,
		Messages: []Message{
			{
				Role:    RoleSystem,
				Content: "Write a headline for the text that the user provides",
			},
			{
				Role:    RoleUser,
				Content: text,
			},
		},
	}

	completion, err := chatCompletion(request)
	if err != nil {
		return "", err
	}
	choice := completion.Choices[0]
	return choice.Message.Content, nil
}
