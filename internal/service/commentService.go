package service

import (
	"context"
	"fmt"
	"strings"
	"threads/internal/entity"
	"threads/internal/repo"
	"unicode/utf8"

	"github.com/google/uuid"
)

type CommentService interface {
	CreateComment(ctx context.Context, postId string, parentID *string, author string, text string) (*entity.Comment, error)
}

type commentServiceImpl struct {
	repo repo.Repo
}

func NewCommentService(repo repo.Repo) *commentServiceImpl {
	return &commentServiceImpl{
		repo: repo,
	}
}

const (
	commentMaxLen = 2000
)

func (c *commentServiceImpl) CreateComment(ctx context.Context, postId string, parentID *string, author string, text string) (*entity.Comment, error) {
	author = strings.TrimSpace(author)

	if author == "" {
		return nil, fmt.Errorf("author cannot be empty")
	}
	if text == "" {
		return nil, fmt.Errorf("text cannot be empty")
	}
	if utf8.RuneCountInString(text) > commentMaxLen {
		return nil, fmt.Errorf("comment's text is too long")
	}

	if parentID != nil {
		parentComment, err := c.repo.GetComment(ctx, *parentID)
		if err != nil {
			return nil, err
		}

		if parentComment.PostID != postId {
			return nil, fmt.Errorf("parent comment post ID does not match with provided post ID")
		}
	}

	post, err := c.repo.GetPost(ctx, postId)
	if err != nil {
		return nil, err
	}
	if !post.CommentsAllowed {
		return nil, fmt.Errorf("comments are not allowed")
	}

	commentID := uuid.New().String()
	comment := &entity.Comment{
		ID:       commentID,
		PostID:   postId,
		ParentID: parentID,
		Author:   author,
		Text:     text,
	}

	if _, err := c.repo.StoreComment(ctx, comment); err != nil {
		return nil, err
	}

	return comment, nil
}
