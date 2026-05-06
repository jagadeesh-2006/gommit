package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type GeminiProvider struct {
	APIKey      string
	Model       string
	CommitStyle string
	CustomPrompt string
}

type geminiModelsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func (g *GeminiProvider) FetchModels() ([]string, error) {
	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models?key=%s", g.APIKey)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not reach Gemini — check internet connection")
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 401, 403:
		return nil, fmt.Errorf("invalid API key — check your Gemini key")
	case 429:
		return nil, fmt.Errorf("rate limit exceeded — try again later")
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result geminiModelsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	// filter chat compatible models only
	allowed := []string{"gemini"}
	models := []string{}
	for _, m := range result.Models {
		// name comes as "models/gemini-1.5-flash" — extract just the model ID
		parts := strings.Split(m.Name, "/")
		id := parts[len(parts)-1]
		for _, a := range allowed {
			if strings.Contains(strings.ToLower(id), a) {
				models = append(models, id)
				break
			}
		}
	}

	return models, nil
}

// ---- GenerateCommitMessage ----

type geminiRequest struct {
	Contents []geminiContent `json:"contents"`
}

type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

type geminiPart struct {
	Text string `json:"text"`
}

type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

func (g *GeminiProvider) GenerateCommitMessage(diff string) (string, error) {
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
%s
`, g.CommitStyle, userInstructions, diff)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		g.Model, g.APIKey,
	)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("could not reach Gemini — check internet connection")
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 401, 403:
		return "", fmt.Errorf("invalid API key — run `gommit update` to fix it")
	case 429:
		return "", fmt.Errorf("rate limit exceeded — wait or switch provider")
	case 500, 502, 503:
		return "", fmt.Errorf("Gemini is down — try again later")
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result geminiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no response from Gemini")
	}

	return strings.TrimSpace(result.Candidates[0].Content.Parts[0].Text), nil
}