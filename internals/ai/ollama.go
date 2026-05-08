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

type OllamaProvider struct {
	APIKey       string // not used but kept for interface consistency
	Model        string
	CommitStyle  string
	CustomPrompt string
}

type ollamaModelsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func (o *OllamaProvider) FetchModels() ([]string, error) {
	req, err := http.NewRequest("GET", "http://localhost:11434/api/tags", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama not running — start it with `ollama serve`")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result ollamaModelsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if len(result.Models) == 0 {
		return nil, fmt.Errorf("no models found — pull one with `ollama pull llama3`")
	}

	models := []string{}
	for _, m := range result.Models {
		models = append(models, m.Name)
	}

	return models, nil
}

// ---- GenerateCommitMessage ----

type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
}

func (o *OllamaProvider) GenerateCommitMessage(diff string) (string, error) {
	userInstructions := ""
	if o.CustomPrompt != "" {
		userInstructions = fmt.Sprintf("Additional instructions from user: %s\n", o.CustomPrompt)
	}

	styleGuide := commitstyle.GetStyleGuide(o.CommitStyle)
	style := commitstyle.GetStyle(o.CommitStyle)

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
	%s
	`, o.CommitStyle, styleGuide, styleExamples,userInstructions, diff)

	reqBody := ollamaRequest{
		Model: o.Model,
		Messages: []ollamaMessage{
			{Role: "user", Content: prompt},
		},
		Stream: false, // important — get full response at once
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "http://localhost:11434/api/chat", bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 60 * time.Second} // longer timeout for local models
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama not running — start it with `ollama serve`")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result ollamaResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if result.Message.Content == "" {
		return "", fmt.Errorf("no response from ollama — try a different model")
	}

	return strings.TrimSpace(result.Message.Content), nil
}
