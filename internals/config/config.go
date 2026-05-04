package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Version   string `json:"version"`
	Provider  string `json:"provider"`
	Model     string `json:"model"`
	APIKey    string `json:"api_key"`
	CommitStyle string `json:"commit_style"`
	CustomPrompt string `json:"custom_prompt"`

}

func configPath()( string , error) {
	home , err := os.UserHomeDir() //  userhomedir fn returns the home directory of the current user
	if err != nil {
		return "", err
	}

	return filepath.Join(home,".gommit","config.json"), nil  
}


// for run.go
func Load() (*Config, error) {
	path , err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err:= json.Unmarshal(data,&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// for init.go
func Save(cfg *Config) error {
	path ,err := configPath()
	if err != nil {
		return err
	}
	// Ensure the directory exists
	if err:= os.MkdirAll(filepath.Dir(path),0755); err != nil{    
		return err
	}

	data, err := json.MarshalIndent(cfg,"","	")
	if err != nil {
		return err
	}

	return os.WriteFile(path,data,0600) 

}

// for root.go 
func Exists() bool {
	path , err := configPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}