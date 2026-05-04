package main 

import (
	"fmt"
	// "github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/jagadeesh-2006/gommit/internals/git"
	"github.com/jagadeesh-2006/gommit/internals/ai"
)

func main() {
	
// 	provider := &ai.GroqProvider{
//     APIKey: "key",
//     Model:  "llama3-8b-8192",
// }


// models, err := provider.FetchModels()
// fmt.Println(models)
// if err!=nil {
// 	fmt.Println("Error fetching models:", err)
// 	return
// }
// diff, _ := git.Diff()
// message, err := provider.GenerateCommitMessage(diff)
// fmt.Println(message)
diff, err := git.Diff()
    if err != nil {
        fmt.Println("Error getting diff:", err)
        return
    }
    fmt.Println("Got diff ✓")

    // step 2 - generate message
    provider := &ai.GroqProvider{
        APIKey: "key",
        Model:  "llama-3.3-70b-versatile",
    }

    message, err := provider.GenerateCommitMessage(diff)
    if err != nil {
        fmt.Println("Error generating message:", err)
        return
    }

    fmt.Println("Suggested commit:", message)
}