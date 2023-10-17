package main

import "github.com/larryzhao/rye/commands"

func main() {
	root := commands.NewRootCmd()
	root.Execute()
}
