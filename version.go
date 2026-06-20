package rift

import "fmt"

var (
	Version string
	Build   string
)

func Ver() string {
	return fmt.Sprintf("%s+develop.%s", Version, Build)
}
