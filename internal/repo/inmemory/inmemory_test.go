package inmemory

import (
	"context"
	"testing"
	"threads/internal/entity"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest() *repoImpl {
	return New()
}

func TestRepo_GetPost(t *testing.T) {
	tests := []struct {
		name        string
		id          string
		presetState func(repo *repoImpl)
		expectedRes *entity.Post
		expectedErr string
	}{
		{
			name: "success comments allowed",
			id:   "p1",
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{
					ID: "p1", Author: "A", Title: "T", Content: "C", CommentsAllowed: true,
				})
			},
			expectedRes: &entity.Post{
				ID: "p1", Author: "A", Title: "T", Content: "C", CommentsAllowed: true,
			},
			expectedErr: "",
		},
		{
			name: "success comments disabled",
			id:   "p2",
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{
					ID: "p2", Author: "A2", Title: "T2", Content: "C2", CommentsAllowed: false,
				})
			},
			expectedRes: &entity.Post{
				ID: "p2", Author: "A2", Title: "T2", Content: "C2", CommentsAllowed: false,
			},
			expectedErr: "",
		},
		{
			name:        "not found",
			id:          "p_not_exist",
			presetState: func(repo *repoImpl) {},
			expectedRes: nil,
			expectedErr: "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTest()
			tt.presetState(repo)

			post, err := repo.GetPost(context.Background(), tt.id)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, post)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, post.CreatedAt)
				tt.expectedRes.CreatedAt = post.CreatedAt // Синхронизируем сгенерированное время
				assert.Equal(t, tt.expectedRes, post)
			}
		})
	}
}

func TestRepo_GetComment(t *testing.T) {
	parentID := "c1"

	tests := []struct {
		name        string
		id          string
		presetState func(repo *repoImpl)
		expectedRes *entity.Comment
		expectedErr string
	}{
		{
			name: "success top level",
			id:   "c1",
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p1"})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{
					ID: "c1", PostID: "p1", ParentID: nil, Author: "A", Text: "T",
				})
			},
			expectedRes: &entity.Comment{
				ID: "c1", PostID: "p1", ParentID: nil, Author: "A", Text: "T",
			},
			expectedErr: "",
		},
		{
			name: "success reply",
			id:   "c2",
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p1"})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{
					ID: "c2", PostID: "p1", ParentID: &parentID, Author: "A", Text: "T",
				})
			},
			expectedRes: &entity.Comment{
				ID: "c2", PostID: "p1", ParentID: &parentID, Author: "A", Text: "T",
			},
			expectedErr: "",
		},
		{
			name:        "not found",
			id:          "c_none",
			presetState: func(repo *repoImpl) {},
			expectedRes: nil,
			expectedErr: "comment not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTest()
			tt.presetState(repo)

			comment, err := repo.GetComment(context.Background(), tt.id)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, comment)
			} else {
				require.NoError(t, err)
				assert.NotZero(t, comment.CreatedAt)
				tt.expectedRes.CreatedAt = comment.CreatedAt
				assert.Equal(t, tt.expectedRes, comment)
			}
		})
	}
}

func TestRepo_GetPosts(t *testing.T) {
	tests := []struct {
		name        string
		limit       int
		offset      int
		presetState func(repo *repoImpl)
		expectedRes []*entity.Post
		expectedErr string
	}{
		{
			name:   "success valid pagination (reverse chronological)",
			limit:  10,
			offset: 0,
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p1", Author: "A1", Title: "T1", Content: "C1", CommentsAllowed: true})
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p2", Author: "A2", Title: "T2", Content: "C2", CommentsAllowed: false})
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p3", Author: "A3", Title: "T3", Content: "C3", CommentsAllowed: true})
			},
			expectedRes: []*entity.Post{
				{ID: "p3", Author: "A3", Title: "T3", Content: "C3", CommentsAllowed: true},
				{ID: "p2", Author: "A2", Title: "T2", Content: "C2", CommentsAllowed: false},
				{ID: "p1", Author: "A1", Title: "T1", Content: "C1", CommentsAllowed: true},
			},
			expectedErr: "",
		},
		{
			name:   "success offset within range returning multiple values",
			limit:  2,
			offset: 2,
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p1"})
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p2"})
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p3"})
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p4"})
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p5"})
			},
			expectedRes: []*entity.Post{
				{ID: "p3"},
				{ID: "p2"},
			},
			expectedErr: "",
		},
		{
			name:   "success offset exceeds total",
			limit:  10,
			offset: 1000,
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p1"})
			},
			expectedRes: []*entity.Post{},
			expectedErr: "",
		},
		{
			name:   "success limit zero",
			limit:  0,
			offset: 0,
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p1"})
			},
			expectedRes: []*entity.Post{},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTest()
			tt.presetState(repo)

			posts, err := repo.GetPosts(context.Background(), tt.limit, tt.offset)

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

func TestRepo_StorePost(t *testing.T) {
	tests := []struct {
		name        string
		inputPost   *entity.Post
		presetState func(repo *repoImpl)
		expectedErr string
	}{
		{
			name:        "success comments allowed",
			inputPost:   &entity.Post{ID: "p1", Author: "A", Title: "T", Content: "C", CommentsAllowed: true},
			presetState: func(repo *repoImpl) {},
			expectedErr: "",
		},
		{
			name:        "success comments disabled",
			inputPost:   &entity.Post{ID: "p2", Author: "A2", Title: "T2", Content: "C2", CommentsAllowed: false},
			presetState: func(repo *repoImpl) {},
			expectedErr: "",
		},
		{
			name:        "error missing id",
			inputPost:   &entity.Post{ID: "", Author: "A", Title: "T", Content: "C"},
			presetState: func(repo *repoImpl) {},
			expectedErr: "post ID cannot be empty",
		},
		{
			name:      "error duplicate id",
			inputPost: &entity.Post{ID: "p1", Author: "A", Title: "T", Content: "C"},
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p1"})
			},
			expectedErr: "post already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTest()
			tt.presetState(repo)

			post, err := repo.StorePost(context.Background(), tt.inputPost)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.inputPost.ID, post.ID)
				assert.Equal(t, tt.inputPost.Author, post.Author)
				assert.NotZero(t, post.CreatedAt)
			}
		})
	}
}

func TestRepo_StoreComment(t *testing.T) {
	parentID := "c1"

	tests := []struct {
		name         string
		inputComment *entity.Comment
		presetState  func(repo *repoImpl)
		expectedErr  string
	}{
		{
			name:         "success top level",
			inputComment: &entity.Comment{ID: "c1", PostID: "p1", ParentID: nil, Author: "A", Text: "T"},
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p1", CommentsAllowed: true})
			},
			expectedErr: "",
		},
		{
			name:         "success reply",
			inputComment: &entity.Comment{ID: "c2", PostID: "p1", ParentID: &parentID, Author: "A", Text: "T"},
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p1", CommentsAllowed: true})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "c1", PostID: "p1"})
			},
			expectedErr: "",
		},
		{
			name:         "error empty comment ID",
			inputComment: &entity.Comment{ID: "", PostID: "p1", ParentID: nil, Author: "A", Text: "T"},
			presetState:  func(repo *repoImpl) {},
			expectedErr:  "comment ID cannot be empty",
		},
		{
			name:         "error empty post ID",
			inputComment: &entity.Comment{ID: "c1", PostID: "", ParentID: nil, Author: "A", Text: "T"},
			presetState:  func(repo *repoImpl) {},
			expectedErr:  "comment post ID cannot be empty",
		},
		{
			name:         "error duplicate comment ID",
			inputComment: &entity.Comment{ID: "c1", PostID: "p1", ParentID: nil, Author: "A", Text: "T"},
			presetState: func(repo *repoImpl) {
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "c1", PostID: "p1"})
			},
			expectedErr: "comment already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTest()
			tt.presetState(repo)

			comment, err := repo.StoreComment(context.Background(), tt.inputComment)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, comment)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.inputComment.ID, comment.ID)
				assert.NotZero(t, comment.CreatedAt)
			}
		})
	}
}

func TestRepo_UpdateCommentsAllowed(t *testing.T) {
	tests := []struct {
		name        string
		postID      string
		allowed     bool
		presetState func(repo *repoImpl)
		expectedErr string
	}{
		{
			name:    "success false",
			postID:  "p1",
			allowed: false,
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p1", CommentsAllowed: true})
			},
			expectedErr: "",
		},
		{
			name:    "success true",
			postID:  "p1",
			allowed: true,
			presetState: func(repo *repoImpl) {
				_, _ = repo.StorePost(context.Background(), &entity.Post{ID: "p1", CommentsAllowed: false})
			},
			expectedErr: "",
		},
		{
			name:        "not found",
			postID:      "p2",
			allowed:     true,
			presetState: func(repo *repoImpl) {},
			expectedErr: "post not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTest()
			tt.presetState(repo)

			post, err := repo.UpdateCommentsAllowed(context.Background(), tt.postID, tt.allowed)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.allowed, post.CommentsAllowed)
			}
		})
	}
}

func TestRepo_GetCommentsByPostIDs(t *testing.T) {
	tests := []struct {
		name        string
		keys        []entity.ParamKey
		presetState func(repo *repoImpl)
		expectedRes map[entity.ParamKey][]*entity.Comment
		expectedErr string
	}{
		{
			name: "success valid pagination multiple keys",
			keys: []entity.ParamKey{
				{Id: "p1", Limit: 10, Offset: 0},
				{Id: "p2", Limit: 5, Offset: 0},
			},
			presetState: func(repo *repoImpl) {
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "c1", PostID: "p1", Author: "A1"})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "c2", PostID: "p1", Author: "A2"})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "c3", PostID: "p1", Author: "A3"})

				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "c4", PostID: "p2", Author: "A4"})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "c5", PostID: "p2", Author: "A5"})
			},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: "p1", Limit: 10, Offset: 0}: {
					{ID: "c3", PostID: "p1", Author: "A3"},
					{ID: "c2", PostID: "p1", Author: "A2"},
					{ID: "c1", PostID: "p1", Author: "A1"},
				},
				{Id: "p2", Limit: 5, Offset: 0}: {
					{ID: "c5", PostID: "p2", Author: "A5"},
					{ID: "c4", PostID: "p2", Author: "A4"},
				},
			},
			expectedErr: "",
		},
		{
			name: "success offset within range",
			keys: []entity.ParamKey{{Id: "p1", Limit: 2, Offset: 1}},
			presetState: func(repo *repoImpl) {
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "c1", PostID: "p1", Author: "A1"})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "c2", PostID: "p1", Author: "A2"})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "c3", PostID: "p1", Author: "A3"})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "c4", PostID: "p1", Author: "A4"})
			},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: "p1", Limit: 2, Offset: 1}: {
					{ID: "c3", PostID: "p1", Author: "A3"},
					{ID: "c2", PostID: "p1", Author: "A2"},
				},
			},
			expectedErr: "",
		},
		{
			name:        "success offset exceeds total",
			keys:        []entity.ParamKey{{Id: "p1", Limit: 10, Offset: 1000}},
			presetState: func(repo *repoImpl) {},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: "p1", Limit: 10, Offset: 1000}: {},
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTest()
			tt.presetState(repo)

			res, err := repo.GetCommentsByPostIDs(context.Background(), tt.keys)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tt.expectedRes), len(res))

				for k, expectedSlice := range tt.expectedRes {
					actualSlice := res[k]
					require.Equal(t, len(expectedSlice), len(actualSlice))

					for i := range actualSlice {
						assert.NotZero(t, actualSlice[i].CreatedAt)
						expectedSlice[i].CreatedAt = actualSlice[i].CreatedAt
					}
					assert.Equal(t, expectedSlice, actualSlice)
				}
			}
		})
	}
}

func TestRepo_GetRepliesByParentIDs(t *testing.T) {
	parentID1 := "c1"
	parentID2 := "c2"

	tests := []struct {
		name        string
		keys        []entity.ParamKey
		presetState func(repo *repoImpl)
		expectedRes map[entity.ParamKey][]*entity.Comment
		expectedErr string
	}{
		{
			name: "success valid pagination with multiple parents",
			keys: []entity.ParamKey{
				{Id: parentID1, Limit: 10, Offset: 0},
				{Id: parentID2, Limit: 5, Offset: 0},
			},
			presetState: func(repo *repoImpl) {
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "r1", PostID: "p1", ParentID: &parentID1, Author: "A1"})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "r2", PostID: "p1", ParentID: &parentID1, Author: "A2"})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "r3", PostID: "p1", ParentID: &parentID1, Author: "A3"})

				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "r4", PostID: "p1", ParentID: &parentID2, Author: "A4"})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "r5", PostID: "p1", ParentID: &parentID2, Author: "A5"})
			},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: parentID1, Limit: 10, Offset: 0}: {
					{ID: "r3", PostID: "p1", ParentID: &parentID1, Author: "A3"},
					{ID: "r2", PostID: "p1", ParentID: &parentID1, Author: "A2"},
					{ID: "r1", PostID: "p1", ParentID: &parentID1, Author: "A1"},
				},
				{Id: parentID2, Limit: 5, Offset: 0}: {
					{ID: "r5", PostID: "p1", ParentID: &parentID2, Author: "A5"},
					{ID: "r4", PostID: "p1", ParentID: &parentID2, Author: "A4"},
				},
			},
			expectedErr: "",
		},
		{
			name: "success offset within range",
			keys: []entity.ParamKey{{Id: parentID1, Limit: 2, Offset: 1}},
			presetState: func(repo *repoImpl) {
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "r1", PostID: "p1", ParentID: &parentID1})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "r2", PostID: "p1", ParentID: &parentID1})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "r3", PostID: "p1", ParentID: &parentID1})
				_, _ = repo.StoreComment(context.Background(), &entity.Comment{ID: "r4", PostID: "p1", ParentID: &parentID1})
			},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: parentID1, Limit: 2, Offset: 1}: {
					{ID: "r3", PostID: "p1", ParentID: &parentID1},
					{ID: "r2", PostID: "p1", ParentID: &parentID1},
				},
			},
			expectedErr: "",
		},
		{
			name:        "success offset exceeds total",
			keys:        []entity.ParamKey{{Id: parentID1, Limit: 5, Offset: 1000}},
			presetState: func(repo *repoImpl) {},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: parentID1, Limit: 5, Offset: 1000}: {},
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := setupTest()
			tt.presetState(repo)

			res, err := repo.GetRepliesByParentIDs(context.Background(), tt.keys)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, res)
			} else {
				require.NoError(t, err)
				require.Equal(t, len(tt.expectedRes), len(res))

				for k, expectedSlice := range tt.expectedRes {
					actualSlice := res[k]
					require.Equal(t, len(expectedSlice), len(actualSlice))

					for i := range actualSlice {
						assert.NotZero(t, actualSlice[i].CreatedAt)
						expectedSlice[i].CreatedAt = actualSlice[i].CreatedAt
					}
					assert.Equal(t, expectedSlice, actualSlice)
				}
			}
		})
	}
}
