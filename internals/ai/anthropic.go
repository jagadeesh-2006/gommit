package ai

import (
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "bytes"
)

type AnthropicProvider struct {
    APIKey string
    Model  string
    CommitStyle string
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

    client := &http.Client{}
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

    models := []string{}
    for _, m := range result.Data {
        models = append(models, m.ID)
    }

    return models, nil
}

type anthropicRequest struct {
    Model     string              `json:"model"`
    MaxTokens int                 `json:"max_tokens"`
    Messages  []anthropicMessage  `json:"messages"`
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
    prompt := fmt.Sprintf(`You are a git commit message generator.
Given the following git diff, generate a concise commit message in conventional commits format.
Only return the commit message, nothing else.

Commit style: %s
Custom prompt: %s
Git diff:
%s`, a.CommitStyle, a.CustomPrompt, diff)

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

    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return "", err
    }
    defer resp.Body.Close()

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