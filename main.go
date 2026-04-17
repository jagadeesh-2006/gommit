package main 

import (
	"fmt"
	"github.com/jagadeesh-2006/gommit/internals/config"
)

func main() {
	
	// quick test in main.go temporarily
	cfg := &config.Config{
		Version:     "1",
		Provider:    "anthropic",
		Model:       "claude-3-5-sonnet",
		APIKey:      "test-key",
		CommitStyle: "conventional",
	}
	config.Save(cfg)

	loaded, _ := config.Load()
	fmt.Println(loaded.Provider) // should print "anthropic"
}