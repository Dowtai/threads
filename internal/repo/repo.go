package repo

import (
	"context"
	"threads/internal/entity"
)

type Repo interface {
	GetPost(ctx context.Context, id string) (*entity.Post, error)
	GetComment(ctx context.Context, id string) (*entity.Comment, error)
	GetPosts(ctx context.Context, limit, offset int) ([]*entity.Post, error)
	GetPostComments(ctx context.Context, id string, limit, offset int) ([]*entity.Comment, error)
	GetCommentReplies(ctx context.Context, id string, limit, offset int) ([]*entity.Comment, error)
	StorePost(ctx context.Context, post *entity.Post) (*entity.Post, error)
	StoreComment(ctx context.Context, comment *entity.Comment) (*entity.Comment, error)
	UpdateCommentsAllowed(ctx context.Context, postID string, allowed bool) (*entity.Post, error)
	GetCommentsByPostIDs(ctx context.Context, ids []string) (map[string][]*entity.Comment, error)
	GetRepliesByParentIDs(ctx context.Context, ids []string) (map[string][]*entity.Comment, error)
}
