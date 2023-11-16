package repo

import (
	"context"
)

type Context struct {
	context.Context
	Repo *Repo
}

func NewContext(repo *Repo) *Context {
	ctx := context.Background()

	return &Context{
		ctx,
		repo,
	}
}
