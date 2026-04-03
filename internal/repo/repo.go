package repo

import (
	"context"
	"threads/internal/entity"
)

type Repo interface {
	GetPost(ctx context.Context, id string) (*entity.Post, error)
	GetComment(ctx context.Context, id string) (*entity.Comment, error)
	GetPosts(ctx context.Context, limit, offset int) ([]*entity.Post, error)
	StorePost(ctx context.Context, post *entity.Post) (*entity.Post, error)
	StoreComment(ctx context.Context, comment *entity.Comment) (*entity.Comment, error)
	UpdateCommentsAllowed(ctx context.Context, postID string, allowed bool) (*entity.Post, error)
	GetCommentsByPostIDs(ctx context.Context, keys []entity.ParamKey) (map[entity.ParamKey][]*entity.Comment, error)
	GetRepliesByParentIDs(ctx context.Context, keys []entity.ParamKey) (map[entity.ParamKey][]*entity.Comment, error)
}
