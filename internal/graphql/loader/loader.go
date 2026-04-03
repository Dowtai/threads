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
	CommentsByPostID  *dataloader.Loader[string, []*entity.Comment]
	RepliesByParentID *dataloader.Loader[string, []*entity.Comment]
}

func Middleware(repo repo.Repo, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), loadersKey, &Loaders{
			CommentsByPostID: dataloader.NewBatchedLoader(func(ctx context.Context, postIDs []string) []*dataloader.Result[[]*entity.Comment] {
				comments, err := repo.GetCommentsByPostIDs(ctx, postIDs)

				results := make([]*dataloader.Result[[]*entity.Comment], len(postIDs))
				for i, id := range postIDs {
					if err != nil {
						results[i] = &dataloader.Result[[]*entity.Comment]{Error: err}
					} else {
						results[i] = &dataloader.Result[[]*entity.Comment]{Data: comments[id]}
					}
				}
				return results
			}),
			RepliesByParentID: dataloader.NewBatchedLoader(func(ctx context.Context, parentIDs []string) []*dataloader.Result[[]*entity.Comment] {
				comments, err := repo.GetRepliesByParentIDs(ctx, parentIDs)

				results := make([]*dataloader.Result[[]*entity.Comment], len(parentIDs))
				for i, id := range parentIDs {
					if err != nil {
						results[i] = &dataloader.Result[[]*entity.Comment]{Error: err}
					} else {
						results[i] = &dataloader.Result[[]*entity.Comment]{Data: comments[id]}
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
