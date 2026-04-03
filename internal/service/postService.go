package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"threads/internal/entity"
	"threads/internal/repo"

	"github.com/google/uuid"
)

type PostService interface {
	CreatePost(ctx context.Context, author string, title string, content string, commentsAllowed bool) (*entity.Post, error)
	Post(ctx context.Context, id string) (*entity.Post, error)
	UpdateCommentsAllowed(ctx context.Context, postID string, author string, allowed bool) (*entity.Post, error)
	Posts(ctx context.Context, limit int32, offset int32) ([]*entity.Post, error)
}

type postServiceImpl struct {
	repo repo.Repo
}

func NewPostService(repo repo.Repo) *postServiceImpl {
	return &postServiceImpl{
		repo: repo,
	}
}

func (p *postServiceImpl) CreatePost(ctx context.Context, author string, title string, content string, commentsAllowed bool) (*entity.Post, error) {
	author = strings.TrimSpace(author)
	title = strings.TrimSpace(title)

	if author == "" {
		return nil, errors.New("author cannot be empty")
	}

	if title == "" {
		return nil, errors.New("title cannot be empty")
	}

	post := &entity.Post{
		ID:              uuid.New().String(),
		Author:          author,
		Title:           title,
		Content:         content,
		CommentsAllowed: commentsAllowed,
	}

	if _, err := p.repo.StorePost(ctx, post); err != nil {
		return nil, err
	}

	return post, nil
}

func (p *postServiceImpl) Post(ctx context.Context, id string) (*entity.Post, error) {
	post, err := p.repo.GetPost(ctx, id)
	if err != nil {
		return nil, err
	}

	return post, nil
}

func (p *postServiceImpl) Posts(ctx context.Context, limit int32, offset int32) ([]*entity.Post, error) {
	if limit < 0 {
		return nil, errors.New("limit cannot be less than 0")
	}
	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be less than 0")
	}

	posts, err := p.repo.GetPosts(ctx, int(limit), int(offset))
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func (p *postServiceImpl) UpdateCommentsAllowed(ctx context.Context, postID string, author string, allowed bool) (*entity.Post, error) {
	post, err := p.Post(ctx, postID)
	if err != nil {
		return nil, errors.New("post not found")
	}
	if post.Author != author {
		return nil, errors.New("author doesn't match with provided author")
	}

	post, err = p.repo.UpdateCommentsAllowed(ctx, postID, allowed)
	if err != nil {
		return nil, err
	}

	return post, nil
}
