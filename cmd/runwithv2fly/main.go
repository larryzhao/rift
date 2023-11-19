package main

import (
	"fmt"
	"os"
	"path"
)

func main() {
	errFile, err := os.OpenFile(path.Join("/Users/larry/.rye", "error.log"), os.O_RDWR|os.O_APPEND, os.ModeAppend)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer errFile.Close()

	outFile, err := os.OpenFile(path.Join("/Users/larry/.rye", "out.log"), os.O_RDWR|os.O_APPEND, os.ModeAppend)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
	defer outFile.Close()

	return
}
