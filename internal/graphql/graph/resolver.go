package graph

import (
	"threads/internal/broker"
	"threads/internal/repo"
	"threads/internal/service"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	PostService    service.PostService
	CommentService service.CommentService
	Broker         broker.CommentBroker
}

func New(repo repo.Repo, broker broker.CommentBroker) *Resolver {
	return &Resolver{
		PostService:    service.NewPostService(repo),
		CommentService: service.NewCommentService(repo),
		Broker:         broker,
	}
}
