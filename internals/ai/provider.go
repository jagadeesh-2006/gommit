package ai

type Provider interface {  
	// Models returns a list of available models for the provider
	Models(apiKey string ) ([]string, error)   // 
	// GenCommitMsg generates a commit message based on the provided diff and returns it as a string
	GenCommitMsg(diff string) (string, error)
}	