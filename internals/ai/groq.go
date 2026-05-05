package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type GroqProvider struct {
	APIKey string
	Model  string
	CommitStyle string
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

	var result groqModelsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	models := []string{}
	for _, m := range result.Data {
		models = append(models, m.ID)
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

func (g *GroqProvider) GenerateCommitMessage(diff string) (string, error) {
	prompt := fmt.Sprintf(`You are a git commit message generator.
Given the following git diff, generate a concise commit message in conventional commits format.
Only return the commit message, nothing else.

Commit style: %s
Custom prompt: %s
Git diff:
%s`, g.CommitStyle, g.CustomPrompt, diff)

	reqBody := groqRequest{
		Model: g.Model,
		Messages: []groqMessage{
			{Role: "user", Content: prompt},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(bodyBytes))
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
	var result groqResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from groq")
	}

	return result.Choices[0].Message.Content, nil
}