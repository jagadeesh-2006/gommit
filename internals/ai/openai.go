package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
	"strings"
)

type OpenAIProvider struct {
	APIKey string
	Model  string
	CommitStyle string
	CustomPrompt string
}

type openAIModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

func (g *OpenAIProvider) FetchModels() ([]string, error) {
	req, err := http.NewRequest("GET", "https://api.openai.com/v1/models", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+g.APIKey)

	client := &http.Client{
		Timeout: 15* time.Second,
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

	var result openAIModelsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	allowed := []string{"gpt-3", "gpt-4", "o1", "o3"}

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

type openAIRequest struct {
	Model    string        `json:"model"`
	Messages []openAIMessage `json:"messages"`
}

type openAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (g *OpenAIProvider) GenerateCommitMessage(diff string) (string, error) {
	userInstructions := ""
	if g.CustomPrompt != "" {
		userInstructions = fmt.Sprintf("Additional instructions from user: %s", g.CustomPrompt)
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

	Commit style guide:
	- conventional: feat(scope): description  or  fix(scope): description
	- simple: short description of what changed
	- emoji: ✨ description  or  🐛 description  or  📝 description

	Git diff:
	%s`, g.CommitStyle, userInstructions, diff)

	reqBody := openAIRequest{
		Model: g.Model,
		Messages: []openAIMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/chat/completions", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}

	req.Header.Set("Authorization", "Bearer "+g.APIKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 15 * time.Second,
	}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
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

	var result openAIResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from openai")
	}

	return result.Choices[0].Message.Content, nil
}