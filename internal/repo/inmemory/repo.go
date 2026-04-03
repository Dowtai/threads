package inmemory

import (
	"context"
	"errors"
	"sync"
	"threads/internal/entity"
	"time"
)

type repoImpl struct {
	mu sync.RWMutex

	posts    map[string]*entity.Post
	comments map[string]*entity.Comment

	postsId []string

	postComments   map[string][]string
	commentReplies map[string][]string
}

func New() *repoImpl {
	return &repoImpl{
		mu:             sync.RWMutex{},
		posts:          make(map[string]*entity.Post),
		comments:       make(map[string]*entity.Comment),
		postsId:        []string{},
		postComments:   make(map[string][]string),
		commentReplies: make(map[string][]string),
	}
}

func (r *repoImpl) GetPost(_ context.Context, id string) (*entity.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if post, ok := r.posts[id]; ok {
		postCopy := *post
		return &postCopy, nil
	}

	return nil, errors.New("post not found")
}

func (r *repoImpl) GetComment(_ context.Context, id string) (*entity.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if comment, ok := r.comments[id]; ok {
		commentCopy := *comment
		return &commentCopy, nil
	}

	return nil, errors.New("comment not found")
}

func (r *repoImpl) GetPosts(_ context.Context, limit, offset int) ([]*entity.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	total := len(r.postsId)

	if offset >= total {
		offset = total
		// ну или возвращать ошибку тут, или возвращать пустой список
		// в целом мне кажется, что оба варианта норм
	}

	limit = min(limit, total-offset)
	posts := make([]*entity.Post, limit)

	for i := 0; i < limit; i++ {
		realI := total - 1 - offset - i
		id := r.postsId[realI]
		post := *r.posts[id]
		posts[i] = &post
	}

	return posts, nil
}

func (r *repoImpl) StorePost(_ context.Context, post *entity.Post) (*entity.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if post.ID == "" {
		return nil, errors.New("post ID cannot be empty")
	}

	if _, ok := r.posts[post.ID]; ok {
		return nil, errors.New("post already exists")
	}

	post.CreatedAt = time.Now().UTC()
	r.posts[post.ID] = post
	r.postsId = append(r.postsId, post.ID)

	return post, nil
}

func (r *repoImpl) StoreComment(_ context.Context, comment *entity.Comment) (*entity.Comment, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if comment.ID == "" {
		return nil, errors.New("comment ID cannot be empty")
	}
	if comment.PostID == "" {
		return nil, errors.New("comment post ID cannot be empty")
	}

	if _, ok := r.comments[comment.ID]; ok {
		return nil, errors.New("comment already exists")
	}

	comment.CreatedAt = time.Now().UTC()
	r.comments[comment.ID] = comment

	if comment.ParentID != nil {
		r.commentReplies[*comment.ParentID] = append(r.commentReplies[*comment.ParentID], comment.ID)
	} else {
		r.postComments[comment.PostID] = append(r.postComments[comment.PostID], comment.ID)
	}

	return comment, nil
}

func (r *repoImpl) UpdateCommentsAllowed(_ context.Context, postID string, allowed bool) (*entity.Post, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.posts[postID]; !ok {
		return nil, errors.New("post not found")
	}

	post := r.posts[postID]
	post.CommentsAllowed = allowed
	postCopy := *post

	return &postCopy, nil
}

func (r *repoImpl) GetCommentsByPostIDs(_ context.Context, keys []entity.ParamKey) (map[entity.ParamKey][]*entity.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[entity.ParamKey][]*entity.Comment)
	for _, key := range keys {
		limit := key.Limit
		offset := key.Offset

		postComments := r.postComments[key.Id]
		total := len(postComments)

		if offset >= total {
			offset = total
		}
		limit = min(limit, total-offset)

		result[key] = make([]*entity.Comment, limit)

		for i := 0; i < limit; i++ {
			realI := total - 1 - offset - i
			cID := postComments[realI]
			if comment, ok := r.comments[cID]; ok {
				commentCopy := *comment
				result[key][i] = &commentCopy
			} else {
				return nil, errors.New("comment was lost")
			}
		}
	}

	return result, nil
}

func (r *repoImpl) GetRepliesByParentIDs(_ context.Context, keys []entity.ParamKey) (map[entity.ParamKey][]*entity.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[entity.ParamKey][]*entity.Comment)
	for _, key := range keys {
		limit := key.Limit
		offset := key.Offset

		commentReplies := r.commentReplies[key.Id]
		total := len(commentReplies)

		if offset >= len(commentReplies) {
			offset = len(commentReplies)
		}
		limit = min(limit, len(commentReplies)-offset)

		result[key] = make([]*entity.Comment, limit)

		for i := 0; i < limit; i++ {
			realI := total - 1 - offset - i
			cID := commentReplies[realI]
			if comment, ok := r.comments[cID]; ok {
				commentCopy := *comment
				result[key][i] = &commentCopy
			} else {
				return nil, errors.New("comment was lost")
			}
		}
	}

	return result, nil
}
