package main 

import (
	"fmt"
	// "github.com/jagadeesh-2006/gommit/internals/config"
	"github.com/jagadeesh-2006/gommit/internals/git"
)

func main() {
	
	diff , err := git.Diff()
	if(err!= nil) {
		fmt.Println("Error running git diff:", err)
		return
	}
	fmt.Println("Git Diff Output:")
	fmt.Println(diff)
	

}