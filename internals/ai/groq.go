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

type GroqProvider struct {
	APIKey       string
	Model        string
	CommitStyle  string
	CustomPrompt string
}

type groqModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (g *GroqProvider) FetchModels() ([]string, error) {
	req, err := http.NewRequest("GET", "https://api.groq.com/openai/v1/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+g.APIKey)

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

	var result groqModelsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	allowed := []string{"llama", "mixtral", "gemma", "qwen", "deepseek"}

	models := []string{}
	for _, m := range result.Data {
		for _, a := range allowed {
			if strings.Contains(strings.ToLower(m.ID), a) {
				models = append(models, m.ID)
				break
			}
		}
	}

	return models, nil
}

type groqRequest struct {
	Model    string        `json:"model"`
	Messages []groqMessage `json:"messages"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (g *GroqProvider) GenerateCommitMessage(diff string, context string, previousMessage string, customPrompt string) ([]string, error) {
	var prompt string

	// If a custom prompt is provided (e.g. from BuildStructuredPrompt), use it directly
	if customPrompt != "" {
		prompt = customPrompt
	} else {
		// Otherwise build the default prompt
		userInstructions := ""
		if g.CustomPrompt != "" {
			userInstructions = fmt.Sprintf("Additional instructions from user: %s\n", g.CustomPrompt)
		}

		contextInfo := ""
		if context != "" {
			contextInfo = fmt.Sprintf("Context about changes: %s\n", context)
		}

		previousSection := ""
		if previousMessage != "" {
			previousSection = fmt.Sprintf("Previous message (do NOT repeat or rephrase this): %s\nTry a completely different angle.\n", previousMessage)
		}

		styleGuide := commitstyle.GetStyleGuide(g.CommitStyle)
		style := commitstyle.GetStyle(g.CommitStyle)

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
===END DIFF===`, previousSection, g.CommitStyle, styleGuide, styleExamples, userInstructions, contextInfo, diff)
	}

	reqBody := groqRequest{
		Model: g.Model,
		Messages: []groqMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+g.APIKey)
	req.Header.Set("Content-Type", "application/json")

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
	var result groqResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no response from groq")
	}

	messages := parseMessages(result.Choices[0].Message.Content)
	if len(messages) == 0 {
		return nil, fmt.Errorf("failed to parse commit messages")
	}

	return messages, nil
}
