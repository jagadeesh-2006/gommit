package ai

import (
    "encoding/json"
	"fmt"
	"io"
	"net/http"
	"bytes"
	"time"
	"strings"
	"github.com/jagadeesh-2006/gommit/internals/commitstyle"
)

type AnthropicProvider struct {
	APIKey       string
	Model        string
	CommitStyle  string
	CustomPrompt string
}

type anthropicModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (a *AnthropicProvider) FetchModels() ([]string, error) {
	req, err := http.NewRequest("GET", "https://api.anthropic.com/v1/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", a.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result anthropicModelsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	blocked := []string{"instant", "legacy"}

	models := []string{}
	for _, m := range result.Data {
		isBlocked := false
		for _, b := range blocked {
			if strings.Contains(strings.ToLower(m.ID), b) {
				isBlocked = true
				break
			}
		}
		if !isBlocked {
			models = append(models, m.ID)
		}
	}
	return models, nil
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Text string `json:"text"`
	} `json:"content"`
}

func (a *AnthropicProvider) GenerateCommitMessage(diff string) (string, error) {
	userInstructions := ""
	if a.CustomPrompt != "" {
		userInstructions = fmt.Sprintf("Additional instructions from user: %s\n", a.CustomPrompt)
	}

	styleGuide := commitstyle.GetStyleGuide(a.CommitStyle)
	style := commitstyle.GetStyle(a.CommitStyle)

	styleExamples := ""
	if len(style.Examples) > 0 {
		styleExamples = "Examples:\n"
		for _, example := range style.Examples {
			styleExamples += fmt.Sprintf("  - %s\n", example)
		}
	}

	prompt := fmt.Sprintf(`You are an expert git commit message generator.

    Your job is to analyze the given git diff and generate a single, concise commit message.

    Rules:
    - Only return the commit message, nothing else
    - Be specific about what changed, not just that something changed
    - Focus on WHY the change was made if it's clear from the diff
    - Keep it under 100 characters

    Commit style: %s
    %s
    %s
    %s
    Git diff:
    %s`, a.CommitStyle, styleGuide, styleExamples,userInstructions, diff)

	reqBody := anthropicRequest{
		Model:     a.Model,
		MaxTokens: 256,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("x-api-key", a.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// response status for failing request
	switch resp.StatusCode {
	case 429:
		return "", fmt.Errorf("rate limit exceeded: %s", resp.Status)
	case 401:
		return "", fmt.Errorf("invalid API key - run `gommit update` to set valid key")
	case 500, 502, 503:
		return "", fmt.Errorf("provider is down try again later")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result anthropicResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Content) == 0 {
		return "", fmt.Errorf("no response from anthropic")
	}

	return result.Content[0].Text, nil
}
