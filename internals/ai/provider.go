package ai

type Provider interface {  
	FetchModels() ([]string, error)   // Fetch available models
	GenerateCommitMessage(diff string) (string, error)
}	

func GetProvider(name string, apiKey string, model string, style string) Provider {
	switch name {
		case "anthropic":
			return &AnthropicProvider{
				APIKey: apiKey,
				Model:  model,
				CommitStyle:  style,
			}
			
		case "groq":
			return &GroqProvider{
				APIKey: apiKey,
				Model:  model,
				CommitStyle:  style,
			}

		case "openai":
			return &OpenAIProvider{
				APIKey: apiKey,
				Model:  model,
				CommitStyle:  style,
			}
		default:
			return nil
	}
}