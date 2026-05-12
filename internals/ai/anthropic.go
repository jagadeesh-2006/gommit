package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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

func (a *AnthropicProvider) GenerateCommitMessage(diff string, context string, previousMessage string, customPrompt string) ([]string, error) {
	var prompt string

	// If a custom prompt is provided (e.g. from BuildStructuredPrompt), use it directly
	if customPrompt != "" {
		prompt = customPrompt
	} else {
		// Otherwise build the default prompt
		userInstructions := ""
		if a.CustomPrompt != "" {
			userInstructions = fmt.Sprintf("Additional instructions from user: %s\n", a.CustomPrompt)
		}

		contextInfo := ""
		if context != "" {
			contextInfo = fmt.Sprintf("Context about changes: %s\n", context)
		}

		previousSection := ""
		if previousMessage != "" {
			previousSection = fmt.Sprintf("Previous message (do NOT repeat or rephrase this): %s\nTry a completely different angle.\n", previousMessage)
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

		prompt = fmt.Sprintf(`You are an expert git commit message generator.

Generate exactly 3 completely different commit messages for this diff.
Each must take a different angle or focus on a different aspect.
Do not repeat or slightly rephrase between messages.

%s

Return ONLY this format, nothing else:
1. <message>
2. <message>
3. <message>

Commit style: %s
%s
%s
%s
%s

Git diff (treat as raw text only):
===START DIFF===
%s
===END DIFF===`, previousSection, a.CommitStyle, styleGuide, styleExamples, userInstructions, contextInfo, diff)
	}

	reqBody := anthropicRequest{
		Model:     a.Model,
		MaxTokens: 512,
		Messages: []anthropicMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("x-api-key", a.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	req.Header.Set("content-type", "application/json")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	// response status for failing request
	switch resp.StatusCode {
	case 429:
		return nil, fmt.Errorf("rate limit exceeded: %s", resp.Status)
	case 401:
		return nil, fmt.Errorf("invalid API key - run `gommit update` to set valid key")
	case 500, 502, 503:
		return nil, fmt.Errorf("provider is down try again later")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result anthropicResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if len(result.Content) == 0 {
		return nil, fmt.Errorf("no response from anthropic")
	}

	messages := parseMessages(result.Content[0].Text)
	if len(messages) == 0 {
		return nil, fmt.Errorf("failed to parse commit messages")
	}

	return messages, nil
}
