package inmemory

import (
	"context"
	"errors"
	"sync"
	"threads/internal/entity"
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

	if offset >= len(r.postsId) {
		offset = len(r.postsId)
		// ну или возвращать ошибку тут, или возвращать пустой список
		// в целом мне кажется, что оба варианта норм
	}

	limit = min(limit, len(r.postsId)-offset)
	posts := make([]*entity.Post, limit)
	ids := r.postsId[offset : offset+limit]

	for i, id := range ids {
		post := *r.posts[id]
		posts[i] = &post
	}

	return posts, nil
}

func (r *repoImpl) GetPostComments(_ context.Context, id string, limit, offset int) ([]*entity.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commentsIds := r.postComments[id]
	if offset >= len(commentsIds) {
		offset = len(commentsIds)
		// аналогично GetPosts
	}
	limit = min(limit, len(commentsIds)-offset)
	comments := make([]*entity.Comment, limit)
	ids := commentsIds[offset : offset+limit]

	for i, id := range ids {
		comment := *r.comments[id]
		comments[i] = &comment
	}

	return comments, nil
}

func (r *repoImpl) GetCommentReplies(_ context.Context, id string, limit, offset int) ([]*entity.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	commentsIds := r.commentReplies[id]
	if offset >= len(commentsIds) {
		offset = len(commentsIds)
	}
	limit = min(limit, len(commentsIds)-offset)
	comments := make([]*entity.Comment, limit)
	ids := commentsIds[offset : offset+limit]

	for i, id := range ids {
		comment := *r.comments[id]
		comments[i] = &comment
	}

	return comments, nil
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

func (r *repoImpl) GetCommentsByPostIDs(_ context.Context, ids []string) (map[string][]*entity.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]*entity.Comment)
	for _, id := range ids {
		postComments := r.postComments[id]
		for _, cID := range postComments {
			if comment, ok := r.comments[cID]; ok {
				commentCopy := *comment
				result[id] = append(result[id], &commentCopy)
			}
		}
	}

	return result, nil
}

func (r *repoImpl) GetRepliesByParentIDs(_ context.Context, ids []string) (map[string][]*entity.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string][]*entity.Comment)
	for _, id := range ids {
		commentReplies := r.commentReplies[id]
		for _, cID := range commentReplies {
			if comment, ok := r.comments[cID]; ok {
				commentCopy := *comment
				result[id] = append(result[id], &commentCopy)
			}
		}
	}

	return result, nil
}
