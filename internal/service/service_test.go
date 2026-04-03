package service

import (
	"context"
	"strings"
	"testing"
	"threads/internal/entity"
	"threads/internal/repo"
	"threads/internal/repo/inmemory"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupPostServiceTest() (PostService, repo.Repo) {
	r := inmemory.New()
	return NewPostService(r), r
}

func setupCommentServiceTest() (CommentService, repo.Repo) {
	r := inmemory.New()
	return NewCommentService(r), r
}

func TestCommentService_CreateComment(t *testing.T) {
	parentID := "c1"
	otherPostParentID := "c2"

	tests := []struct {
		name        string
		postID      string
		parentID    *string
		author      string
		text        string
		presetState func(r repo.Repo)
		expectedRes *entity.Comment
		expectedErr string
	}{
		{
			name:     "success top level",
			postID:   "p1",
			parentID: nil,
			author:   "A1",
			text:     "Some text",
			presetState: func(r repo.Repo) {
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p1", CommentsAllowed: true})
			},
			expectedRes: &entity.Comment{
				PostID:   "p1",
				ParentID: nil,
				Author:   "A1",
				Text:     "Some text",
			},
			expectedErr: "",
		},
		{
			name:     "success reply",
			postID:   "p1",
			parentID: &parentID,
			author:   " A2 ",
			text:     "Reply text",
			presetState: func(r repo.Repo) {
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p1", CommentsAllowed: true})
				_, _ = r.StoreComment(context.Background(), &entity.Comment{ID: "c1", PostID: "p1"})
			},
			expectedRes: &entity.Comment{
				PostID:   "p1",
				ParentID: &parentID,
				Author:   "A2",
				Text:     "Reply text",
			},
			expectedErr: "",
		},
		{
			name:        "error empty author",
			postID:      "p1",
			parentID:    nil,
			author:      "   ",
			text:        "Some text",
			presetState: func(r repo.Repo) {},
			expectedRes: nil,
			expectedErr: "author cannot be empty",
		},
		{
			name:        "error empty text",
			postID:      "p1",
			parentID:    nil,
			author:      "A1",
			text:        "   ",
			presetState: func(r repo.Repo) {},
			expectedRes: nil,
			expectedErr: "text cannot be empty",
		},
		{
			name:        "error text too long",
			postID:      "p1",
			parentID:    nil,
			author:      "A1",
			text:        strings.Repeat("a", 2001),
			presetState: func(r repo.Repo) {},
			expectedRes: nil,
			expectedErr: "comment's text cannot contain more than 2000 characters",
		},
		{
			name:     "error parent comment not found",
			postID:   "p1",
			parentID: &parentID,
			author:   "A1",
			text:     "Text",
			presetState: func(r repo.Repo) {
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p1", CommentsAllowed: true})
			},
			expectedRes: nil,
			expectedErr: "comment not found",
		},
		{
			name:     "error parent post mismatch",
			postID:   "p1",
			parentID: &otherPostParentID,
			author:   "A1",
			text:     "Some text",
			presetState: func(r repo.Repo) {
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p1", CommentsAllowed: true})
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p2", CommentsAllowed: true})
				_, _ = r.StoreComment(context.Background(), &entity.Comment{ID: "c2", PostID: "p2"})
			},
			expectedRes: nil,
			expectedErr: "parent comment post ID does not match with provided post ID",
		},
		{
			name:        "error post not found",
			postID:      "p_none",
			parentID:    nil,
			author:      "A1",
			text:        "Text",
			presetState: func(r repo.Repo) {},
			expectedRes: nil,
			expectedErr: "post not found",
		},
		{
			name:     "error comments disabled",
			postID:   "p1",
			parentID: nil,
			author:   "A1",
			text:     "Some text",
			presetState: func(r repo.Repo) {
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p1", CommentsAllowed: false})
			},
			expectedRes: nil,
			expectedErr: "comments are not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, r := setupCommentServiceTest()
			tt.presetState(r)

			comment, err := svc.CreateComment(context.Background(), tt.postID, tt.parentID, tt.author, tt.text)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, comment)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, comment.ID)
				assert.NotZero(t, comment.CreatedAt)

				tt.expectedRes.ID = comment.ID
				tt.expectedRes.CreatedAt = comment.CreatedAt
				assert.Equal(t, tt.expectedRes, comment)
			}
		})
	}
}

func TestPostService_CreatePost(t *testing.T) {
	tests := []struct {
		name            string
		author          string
		title           string
		content         string
		commentsAllowed bool
		expectedRes     *entity.Post
		expectedErr     string
	}{
		{
			name:            "success with comments allowed",
			author:          " A1 ",
			title:           " T1 ",
			content:         "C1",
			commentsAllowed: true,
			expectedRes: &entity.Post{
				Author:          "A1",
				Title:           "T1",
				Content:         "C1",
				CommentsAllowed: true,
			},
			expectedErr: "",
		},
		{
			name:            "success empty content",
			author:          "A2",
			title:           "T2",
			content:         "",
			commentsAllowed: false,
			expectedRes: &entity.Post{
				Author:          "A2",
				Title:           "T2",
				Content:         "",
				CommentsAllowed: false,
			},
			expectedErr: "",
		},
		{
			name:            "error empty author",
			author:          "   ",
			title:           "T1",
			content:         "C1",
			commentsAllowed: true,
			expectedRes:     nil,
			expectedErr:     "author cannot be empty",
		},
		{
			name:            "error empty title",
			author:          "A1",
			title:           "   ",
			content:         "C1",
			commentsAllowed: true,
			expectedRes:     nil,
			expectedErr:     "title cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, _ := setupPostServiceTest()

			post, err := svc.CreatePost(context.Background(), tt.author, tt.title, tt.content, tt.commentsAllowed)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, post)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, post.ID)
				assert.NotZero(t, post.CreatedAt)

				tt.expectedRes.ID = post.ID
				tt.expectedRes.CreatedAt = post.CreatedAt
				assert.Equal(t, tt.expectedRes, post)
			}
		})
	}
}

func TestPostService_Post(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		presetState func(r repo.Repo)
		expectedRes *entity.Post
		expectedErr string
	}{
		{
			name: "success",
			id:   "p1",
			presetState: func(r repo.Repo) {
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p1", Author: "A", Title: "T"})
			},
			expectedRes: &entity.Post{
				ID: "p1", Author: "A", Title: "T",
			},
			expectedErr: "",
		},
		{
			name:        "error not found",
			id:          "p_not_exist",
			presetState: func(r repo.Repo) {},
			expectedRes: nil,
			expectedErr: "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, r := setupPostServiceTest()
			tt.presetState(r)

			post, err := svc.Post(context.Background(), tt.id)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, post)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, post.CreatedAt)
				tt.expectedRes.CreatedAt = post.CreatedAt
				assert.Equal(t, tt.expectedRes, post)
			}
		})
	}
}

func TestPostService_Posts(t *testing.T) {
	tests := []struct {
		name        string
		limit       int32
		offset      int32
		presetState func(r repo.Repo)
		expectedRes []*entity.Post
		expectedErr string
	}{
		{
			name:   "success valid pagination",
			limit:  10,
			offset: 0,
			presetState: func(r repo.Repo) {
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p1", Author: "A1"})
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p2", Author: "A2"})
			},
			expectedRes: []*entity.Post{
				{ID: "p2", Author: "A2"},
				{ID: "p1", Author: "A1"},
			},
			expectedErr: "",
		},
		{
			name:   "success offset exceeds total",
			limit:  2,
			offset: 2,
			presetState: func(r repo.Repo) {
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p1", Author: "A1"})
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p2", Author: "A2"})
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p3", Author: "A3"})
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p4", Author: "A4"})
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p5", Author: "A5"})
			},
			expectedRes: []*entity.Post{
				{ID: "p3", Author: "A3"},
				{ID: "p2", Author: "A2"},
			},
			expectedErr: "",
		},
		{
			name:        "error negative limit",
			limit:       -1,
			offset:      0,
			presetState: func(r repo.Repo) {},
			expectedRes: nil,
			expectedErr: "limit cannot be less than 0",
		},
		{
			name:        "error negative offset",
			limit:       10,
			offset:      -5,
			presetState: func(r repo.Repo) {},
			expectedRes: nil,
			expectedErr: "offset cannot be less than 0",
		},
		{
			name:   "success zero limit",
			limit:  0,
			offset: 0,
			presetState: func(r repo.Repo) {
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p1", Author: "A1"})
			},
			expectedRes: []*entity.Post{},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, r := setupPostServiceTest()
			tt.presetState(r)

			posts, err := svc.Posts(context.Background(), tt.limit, tt.offset)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, posts)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tt.expectedRes), len(posts))

				for i := range posts {
					assert.NotZero(t, posts[i].CreatedAt)
					tt.expectedRes[i].CreatedAt = posts[i].CreatedAt
				}
				assert.Equal(t, tt.expectedRes, posts)
			}
		})
	}
}

func TestPostService_UpdateCommentsAllowed(t *testing.T) {
	tests := []struct {
		name        string
		postID      string
		author      string
		allowed     bool
		presetState func(r repo.Repo)
		expectedRes *entity.Post
		expectedErr string
	}{
		{
			name:    "success disable comments",
			postID:  "p1",
			author:  "A1",
			allowed: false,
			presetState: func(r repo.Repo) {
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p1", Author: "A1", CommentsAllowed: true})
			},
			expectedRes: &entity.Post{
				ID: "p1", Author: "A1", CommentsAllowed: false,
			},
			expectedErr: "",
		},
		{
			name:    "success enable comments",
			postID:  "p1",
			author:  "A1",
			allowed: true,
			presetState: func(r repo.Repo) {
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p1", Author: "A1", CommentsAllowed: false})
			},
			expectedRes: &entity.Post{
				ID: "p1", Author: "A1", CommentsAllowed: true,
			},
			expectedErr: "",
		},
		{
			name:        "error post not found",
			postID:      "p_none",
			author:      "A1",
			allowed:     true,
			presetState: func(r repo.Repo) {},
			expectedRes: nil,
			expectedErr: "post not found",
		},
		{
			name:    "error author mismatch",
			postID:  "p1",
			author:  "WrongAuthor",
			allowed: false,
			presetState: func(r repo.Repo) {
				_, _ = r.StorePost(context.Background(), &entity.Post{ID: "p1", Author: "A1", CommentsAllowed: true})
			},
			expectedRes: nil,
			expectedErr: "author doesn't match",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, r := setupPostServiceTest()
			tt.presetState(r)

			post, err := svc.UpdateCommentsAllowed(context.Background(), tt.postID, tt.author, tt.allowed)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, post)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, post.CreatedAt)
				tt.expectedRes.CreatedAt = post.CreatedAt
				assert.Equal(t, tt.expectedRes, post)
			}
		})
	}
}
