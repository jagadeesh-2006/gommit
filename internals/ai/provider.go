package ai

type Provider interface {  
	FetchModels() ([]string, error)   // 
	GenerateCommitMessage(diff string) (string, error)
}	

func GetProvider(name string, apiKey string, model string) Provider {
	switch name {
		case "anthropic":
			provider := &GroqProvider{
			APIKey: "apikey",
			Model:  "model",	
			}
			return provider
		case "groq":
			return &GroqProvider{
				APIKey: apiKey,
				Model:  model,
			}

		case "openai":
			return &OpenAIProvider{
				APIKey: apiKey,
				Model:  model,
			}
		default:
			return nil
	}
}