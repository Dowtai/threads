package loader

import (
	"context"
	"net/http"
	"threads/internal/entity"
	"threads/internal/repo"

	"github.com/graph-gophers/dataloader/v7"
)

type ctxKey string

const loadersKey ctxKey = "dataloaders"

type Loaders struct {
	CommentsByPostID  *dataloader.Loader[entity.ParamKey, []*entity.Comment]
	RepliesByParentID *dataloader.Loader[entity.ParamKey, []*entity.Comment]
}

func Middleware(repo repo.Repo, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), loadersKey, &Loaders{
			CommentsByPostID: dataloader.NewBatchedLoader(func(ctx context.Context, keys []entity.ParamKey) []*dataloader.Result[[]*entity.Comment] {
				comments, err := repo.GetCommentsByPostIDs(ctx, keys)

				results := make([]*dataloader.Result[[]*entity.Comment], len(keys))
				for i, key := range keys {
					if err != nil {
						results[i] = &dataloader.Result[[]*entity.Comment]{Error: err}
					} else {
						results[i] = &dataloader.Result[[]*entity.Comment]{Data: comments[key]}
					}
				}
				return results
			}),
			RepliesByParentID: dataloader.NewBatchedLoader(func(ctx context.Context, keys []entity.ParamKey) []*dataloader.Result[[]*entity.Comment] {
				comments, err := repo.GetRepliesByParentIDs(ctx, keys)

				results := make([]*dataloader.Result[[]*entity.Comment], len(keys))
				for i, key := range keys {
					if err != nil {
						results[i] = &dataloader.Result[[]*entity.Comment]{Error: err}
					} else {
						results[i] = &dataloader.Result[[]*entity.Comment]{Data: comments[key]}
					}
				}
				return results
			}),
		})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(loadersKey).(*Loaders)
}
