package ai

type Provider interface {
	FetchModels() ([]string, error) // Fetch available models
	GenerateCommitMessage(diff string, context string, previousMessage string, customPrompt string) ([]string, error)
}

func GetProvider(name string, apiKey string, model string, style string, customPrompt string) Provider {
	switch name {
	case "anthropic":
		return &AnthropicProvider{
			APIKey:       apiKey,
			Model:        model,
			CommitStyle:  style,
			CustomPrompt: customPrompt,
		}

	case "groq":
		return &GroqProvider{
			APIKey:       apiKey,
			Model:        model,
			CommitStyle:  style,
			CustomPrompt: customPrompt,
		}

	case "openai":
		return &OpenAIProvider{
			APIKey:       apiKey,
			Model:        model,
			CommitStyle:  style,
			CustomPrompt: customPrompt,
		}
	case "ollama":
		return &OllamaProvider{
			APIKey:       apiKey,
			Model:        model,
			CommitStyle:  style,
			CustomPrompt: customPrompt,
		}
	case "gemini":
		return &GeminiProvider{
			APIKey:       apiKey,
			Model:        model,
			CommitStyle:  style,
			CustomPrompt: customPrompt,
		}
	default:
		return nil
	}
}
