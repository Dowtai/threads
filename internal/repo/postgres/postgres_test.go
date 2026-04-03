package postgres

import (
	"context"
	"errors"
	"testing"
	"threads/internal/entity"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/pashagolub/pgxmock/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTest(t *testing.T) (pgxmock.PgxPoolIface, *repoImpl) {
	mock, err := pgxmock.NewPool()
	require.NoError(t, err)
	repo := New(mock)
	return mock, repo
}

func TestRepo_GetPost(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		id          string
		mockSetup   func(mock pgxmock.PgxPoolIface)
		expectedRes *entity.Post
		expectedErr string
	}{
		{
			name: "success comments allowed",
			id:   "p1",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "author", "title", "content", "comments_allowed", "created_at"}).
					AddRow("p1", "Author", "Title", "Content", true, now)
				mock.ExpectQuery("^SELECT id, author").WithArgs("p1").WillReturnRows(rows)
			},
			expectedRes: &entity.Post{
				ID:              "p1",
				Author:          "Author",
				Title:           "Title",
				Content:         "Content",
				CommentsAllowed: true,
				CreatedAt:       now,
			},
			expectedErr: "",
		},
		{
			name: "success comments disabled",
			id:   "p2",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"id", "author", "title", "content", "comments_allowed", "created_at"}).
					AddRow("p2", "Author2", "Title2", "Content2", false, now)
				mock.ExpectQuery("^SELECT id, author").WithArgs("p2").WillReturnRows(rows)
			},
			expectedRes: &entity.Post{
				ID:              "p2",
				Author:          "Author2",
				Title:           "Title2",
				Content:         "Content2",
				CommentsAllowed: false,
				CreatedAt:       now,
			},
			expectedErr: "",
		},
		{
			name: "not found",
			id:   "p_not_exist",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("^SELECT id, author").WithArgs("p_not_exist").WillReturnError(pgx.ErrNoRows)
			},
			expectedRes: nil,
			expectedErr: "post not found",
		},
		{
			name: "db error",
			id:   "p_err",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("^SELECT id, author").WithArgs("p_err").WillReturnError(errors.New("db error"))
			},
			expectedRes: nil,
			expectedErr: "db error",
		},
		{
			name: "context canceled",
			id:   "p1",
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("^SELECT id, author").WithArgs("p1").WillReturnError(context.Canceled)
			},
			expectedRes: nil,
			expectedErr: "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, repo := setupTest(t)
			defer mock.Close()

			tt.mockSetup(mock)
			post, err := repo.GetPost(context.Background(), tt.id)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, post)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepo_GetComment(t *testing.T) {
	mock, repo := setupTest(t)
	defer mock.Close()
	now := time.Now()
	parentID := "c1"

	tests := []struct {
		name        string
		id          string
		mockSetup   func()
		expectedRes *entity.Comment
		expectedErr string
	}{
		{
			name: "success top level",
			id:   "c1",
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "post_id", "parent_id", "author", "text", "created_at"}).
					AddRow("c1", "p1", nil, "Author", "Text", now)
				mock.ExpectQuery("^SELECT id, post_id").WithArgs("c1").WillReturnRows(rows)
			},
			expectedRes: &entity.Comment{
				ID:        "c1",
				PostID:    "p1",
				ParentID:  nil,
				Author:    "Author",
				Text:      "Text",
				CreatedAt: now,
			},
			expectedErr: "",
		},
		{
			name: "success reply",
			id:   "c2",
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "post_id", "parent_id", "author", "text", "created_at"}).
					AddRow("c2", "p1", &parentID, "Author2", "Reply", now)
				mock.ExpectQuery("^SELECT id, post_id").WithArgs("c2").WillReturnRows(rows)
			},
			expectedRes: &entity.Comment{
				ID:        "c2",
				PostID:    "p1",
				ParentID:  &parentID,
				Author:    "Author2",
				Text:      "Reply",
				CreatedAt: now,
			},
			expectedErr: "",
		},
		{
			name: "not found",
			id:   "c_unknown_id",
			mockSetup: func() {
				mock.ExpectQuery("^SELECT id, post_id").WithArgs("c_unknown_id").WillReturnError(pgx.ErrNoRows)
			},
			expectedRes: nil,
			expectedErr: "comment not found",
		},
		{
			name: "db error",
			id:   "c_err",
			mockSetup: func() {
				mock.ExpectQuery("^SELECT id, post_id").WithArgs("c_err").WillReturnError(errors.New("db error"))
			},
			expectedRes: nil,
			expectedErr: "db error",
		},
		{
			name: "context canceled",
			id:   "c1",
			mockSetup: func() {
				mock.ExpectQuery("^SELECT id, post_id").WithArgs("c1").WillReturnError(context.Canceled)
			},
			expectedRes: nil,
			expectedErr: "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			comment, err := repo.GetComment(context.Background(), tt.id)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, comment)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, comment)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepo_GetPosts(t *testing.T) {
	mock, repo := setupTest(t)
	defer mock.Close()
	now := time.Now()

	tests := []struct {
		name        string
		limit       int
		offset      int
		mockSetup   func()
		expectedRes []*entity.Post
		expectedErr string
	}{
		{
			name:   "success valid pagination",
			limit:  10,
			offset: 0,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "author", "title", "content", "comments_allowed", "created_at"}).
					AddRow("p1", "A1", "T1", "C1", true, now).
					AddRow("p2", "A2", "T2", "C2", false, now).
					AddRow("p3", "A3", "T3", "C3", false, now).
					AddRow("p4", "A4", "T4", "C4", false, now).
					AddRow("p5", "A5", "T5", "C5", false, now)
				mock.ExpectQuery("^SELECT id, author").WithArgs(10, 0).WillReturnRows(rows)
			},
			expectedRes: []*entity.Post{
				{ID: "p1", Author: "A1", Title: "T1", Content: "C1", CommentsAllowed: true, CreatedAt: now},
				{ID: "p2", Author: "A2", Title: "T2", Content: "C2", CommentsAllowed: false, CreatedAt: now},
				{ID: "p3", Author: "A3", Title: "T3", Content: "C3", CommentsAllowed: false, CreatedAt: now},
				{ID: "p4", Author: "A4", Title: "T4", Content: "C4", CommentsAllowed: false, CreatedAt: now},
				{ID: "p5", Author: "A5", Title: "T5", Content: "C5", CommentsAllowed: false, CreatedAt: now},
			},
			expectedErr: "",
		},
		{
			name:   "success offset within range",
			limit:  2,
			offset: 2,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "author", "title", "content", "comments_allowed", "created_at"}).
					AddRow("p3", "A3", "T3", "C3", false, now).
					AddRow("p4", "A4", "T4", "C4", false, now)
				mock.ExpectQuery("^SELECT id, author").
					WithArgs(2, 2).WillReturnRows(rows)
			},
			expectedRes: []*entity.Post{
				{ID: "p3", Author: "A3", Title: "T3", Content: "C3", CommentsAllowed: false, CreatedAt: now},
				{ID: "p4", Author: "A4", Title: "T4", Content: "C4", CommentsAllowed: false, CreatedAt: now},
			},
			expectedErr: "",
		},
		{
			name:   "success offset exceeds total",
			limit:  10,
			offset: 1000,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "author", "title", "content", "comments_allowed", "created_at"})
				mock.ExpectQuery("^SELECT id, author").WithArgs(10, 1000).WillReturnRows(rows)
			},
			expectedRes: []*entity.Post{},
			expectedErr: "",
		},
		{
			name:   "db error negative offset",
			limit:  10,
			offset: -1,
			mockSetup: func() {
				mock.ExpectQuery("^SELECT id, author").WithArgs(10, -1).WillReturnError(errors.New("OFFSET must not be negative"))
			},
			expectedRes: nil,
			expectedErr: "OFFSET must not be negative",
		},
		{
			name:   "db internal error",
			limit:  10,
			offset: 0,
			mockSetup: func() {
				mock.ExpectQuery("^SELECT id, author").WithArgs(10, 0).WillReturnError(errors.New("connection failed"))
			},
			expectedRes: nil,
			expectedErr: "connection failed",
		},
		{
			name:   "success limit zero",
			limit:  0,
			offset: 0,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "author", "title", "content", "comments_allowed", "created_at"})
				mock.ExpectQuery("^SELECT id, author").WithArgs(0, 0).WillReturnRows(rows)
			},
			expectedRes: []*entity.Post{},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			posts, err := repo.GetPosts(context.Background(), tt.limit, tt.offset)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, posts)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, posts)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepo_StorePost(t *testing.T) {
	mock, repo := setupTest(t)
	defer mock.Close()
	now := time.Now()

	tests := []struct {
		name        string
		inputPost   *entity.Post
		mockSetup   func()
		expectedRes *entity.Post
		expectedErr string
	}{
		{
			name:      "success comments allowed",
			inputPost: &entity.Post{ID: "p1", Author: "A", Title: "T", Content: "C", CommentsAllowed: true},
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"created_at"}).AddRow(now)
				mock.ExpectQuery("^INSERT INTO posts").
					WithArgs("p1", "A", "T", "C", true).WillReturnRows(rows)
			},
			expectedRes: &entity.Post{
				ID:              "p1",
				Author:          "A",
				Title:           "T",
				Content:         "C",
				CommentsAllowed: true,
				CreatedAt:       now,
			},
			expectedErr: "",
		},
		{
			name:      "success comments disabled",
			inputPost: &entity.Post{ID: "p2", Author: "A2", Title: "T2", Content: "C2", CommentsAllowed: false},
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"created_at"}).AddRow(now)
				mock.ExpectQuery("^INSERT INTO posts").
					WithArgs("p2", "A2", "T2", "C2", false).WillReturnRows(rows)
			},
			expectedRes: &entity.Post{
				ID:              "p2",
				Author:          "A2",
				Title:           "T2",
				Content:         "C2",
				CommentsAllowed: false,
				CreatedAt:       now,
			},
			expectedErr: "",
		},
		{
			name:      "db unique violation",
			inputPost: &entity.Post{ID: "p1", Author: "A", Title: "T", Content: "C", CommentsAllowed: true},
			mockSetup: func() {
				mock.ExpectQuery("^INSERT INTO posts").
					WithArgs("p1", "A", "T", "C", true).
					WillReturnError(errors.New("duplicate key"))
			},
			expectedRes: nil,
			expectedErr: "duplicate key",
		},
		{
			name:      "db internal error",
			inputPost: &entity.Post{ID: "p1", Author: "A", Title: "T", Content: "C", CommentsAllowed: true},
			mockSetup: func() {
				mock.ExpectQuery("^INSERT INTO posts").
					WithArgs("p1", "A", "T", "C", true).
					WillReturnError(errors.New("insert failed"))
			},
			expectedRes: nil,
			expectedErr: "insert failed",
		},
		{
			name:      "context canceled",
			inputPost: &entity.Post{ID: "p1", Author: "A", Title: "T", Content: "C", CommentsAllowed: true},
			mockSetup: func() {
				mock.ExpectQuery("^INSERT INTO posts").
					WithArgs("p1", "A", "T", "C", true).
					WillReturnError(context.Canceled)
			},
			expectedRes: nil,
			expectedErr: "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			post, err := repo.StorePost(context.Background(), tt.inputPost)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, post)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepo_StoreComment(t *testing.T) {
	mock, repo := setupTest(t)
	defer mock.Close()
	now := time.Now()
	parentID := "c1"

	tests := []struct {
		name         string
		inputComment *entity.Comment
		mockSetup    func()
		expectedRes  *entity.Comment
		expectedErr  string
	}{
		{
			name:         "success top level",
			inputComment: &entity.Comment{ID: "c1", PostID: "p1", ParentID: nil, Author: "A", Text: "T"},
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"created_at"}).AddRow(now)
				mock.ExpectQuery("^INSERT INTO comments").
					WithArgs("c1", "p1", (*string)(nil), "A", "T").WillReturnRows(rows)
			},
			expectedRes: &entity.Comment{
				ID:        "c1",
				PostID:    "p1",
				ParentID:  nil,
				Author:    "A",
				Text:      "T",
				CreatedAt: now,
			},
			expectedErr: "",
		},
		{
			name:         "success reply",
			inputComment: &entity.Comment{ID: "c2", PostID: "p1", ParentID: &parentID, Author: "A", Text: "T"},
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"created_at"}).AddRow(now)
				mock.ExpectQuery("^INSERT INTO comments").
					WithArgs("c2", "p1", &parentID, "A", "T").WillReturnRows(rows)
			},
			expectedRes: &entity.Comment{
				ID:        "c2",
				PostID:    "p1",
				ParentID:  &parentID,
				Author:    "A",
				Text:      "T",
				CreatedAt: now,
			},
			expectedErr: "",
		},
		{
			name:         "fk violation",
			inputComment: &entity.Comment{ID: "c1", PostID: "unknown", ParentID: nil, Author: "A", Text: "T"},
			mockSetup: func() {
				mock.ExpectQuery("^INSERT INTO comments").
					WithArgs("c1", "unknown", (*string)(nil), "A", "T").
					WillReturnError(errors.New("foreign key constraint"))
			},
			expectedRes: nil,
			expectedErr: "foreign key constraint",
		},
		{
			name:         "db error",
			inputComment: &entity.Comment{ID: "c1", PostID: "p1", ParentID: nil, Author: "A", Text: "T"},
			mockSetup: func() {
				mock.ExpectQuery("^INSERT INTO comments").
					WithArgs("c1", "p1", (*string)(nil), "A", "T").
					WillReturnError(errors.New("insert error"))
			},
			expectedRes: nil,
			expectedErr: "insert error",
		},
		{
			name:         "context canceled",
			inputComment: &entity.Comment{ID: "c1", PostID: "p1", ParentID: nil, Author: "A", Text: "T"},
			mockSetup: func() {
				mock.ExpectQuery("^INSERT INTO comments").
					WithArgs("c1", "p1", (*string)(nil), "A", "T").
					WillReturnError(context.Canceled)
			},
			expectedRes: nil,
			expectedErr: "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			comment, err := repo.StoreComment(context.Background(), tt.inputComment)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, comment)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, comment)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepo_UpdateCommentsAllowed(t *testing.T) {
	mock, repo := setupTest(t)
	defer mock.Close()
	now := time.Now()

	tests := []struct {
		name        string
		postID      string
		allowed     bool
		mockSetup   func()
		expectedRes *entity.Post
		expectedErr string
	}{
		{
			name:    "success false",
			postID:  "p1",
			allowed: false,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "author", "title", "content", "comments_allowed", "created_at"}).
					AddRow("p1", "A", "T", "C", false, now)
				mock.ExpectQuery("^UPDATE posts").WithArgs(false, "p1").WillReturnRows(rows)
			},
			expectedRes: &entity.Post{
				ID:              "p1",
				Author:          "A",
				Title:           "T",
				Content:         "C",
				CommentsAllowed: false,
				CreatedAt:       now,
			},
			expectedErr: "",
		},
		{
			name:    "success true",
			postID:  "p1",
			allowed: true,
			mockSetup: func() {
				rows := pgxmock.NewRows([]string{"id", "author", "title", "content", "comments_allowed", "created_at"}).
					AddRow("p1", "A", "T", "C", true, now)
				mock.ExpectQuery("^UPDATE posts").WithArgs(true, "p1").WillReturnRows(rows)
			},
			expectedRes: &entity.Post{
				ID:              "p1",
				Author:          "A",
				Title:           "T",
				Content:         "C",
				CommentsAllowed: true,
				CreatedAt:       now,
			},
			expectedErr: "",
		},
		{
			name:    "not found",
			postID:  "p2",
			allowed: true,
			mockSetup: func() {
				mock.ExpectQuery("^UPDATE posts").WithArgs(true, "p2").WillReturnError(pgx.ErrNoRows)
			},
			expectedRes: nil,
			expectedErr: "post not found",
		},
		{
			name:    "db error",
			postID:  "p1",
			allowed: false,
			mockSetup: func() {
				mock.ExpectQuery("^UPDATE posts").WithArgs(false, "p1").WillReturnError(errors.New("update failed"))
			},
			expectedRes: nil,
			expectedErr: "update failed",
		},
		{
			name:    "context canceled",
			postID:  "p1",
			allowed: false,
			mockSetup: func() {
				mock.ExpectQuery("^UPDATE posts").WithArgs(false, "p1").WillReturnError(context.Canceled)
			},
			expectedRes: nil,
			expectedErr: "context canceled",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()
			post, err := repo.UpdateCommentsAllowed(context.Background(), tt.postID, tt.allowed)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, post)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, post)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepo_GetCommentsByPostIDs(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name        string
		keys        []entity.ParamKey
		mockSetup   func(mock pgxmock.PgxPoolIface)
		expectedRes map[entity.ParamKey][]*entity.Comment
		expectedErr string
	}{
		{
			name: "success valid pagination with multiple posts and comments",
			keys: []entity.ParamKey{
				{Id: "p1", Limit: 10, Offset: 0},
				{Id: "p2", Limit: 5, Offset: 0},
			},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"req_id", "limit_val", "offset_val", "id", "post_id", "parent_id", "author", "text", "created_at"}).
					AddRow("p1", 10, 0, "c1", "p1", nil, "A1", "T1", now).
					AddRow("p1", 10, 0, "c2", "p1", nil, "A2", "T2", now).
					AddRow("p1", 10, 0, "c3", "p1", nil, "A3", "T3", now).
					AddRow("p2", 5, 0, "c4", "p2", nil, "A4", "T4", now).
					AddRow("p2", 5, 0, "c5", "p2", nil, "A5", "T5", now)
				mock.ExpectQuery("^SELECT.*FROM unnest").
					WithArgs([]string{"p1", "p2"}, []int{10, 5}, []int{0, 0}).WillReturnRows(rows)
			},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: "p1", Limit: 10, Offset: 0}: {
					{ID: "c1", PostID: "p1", ParentID: nil, Author: "A1", Text: "T1", CreatedAt: now},
					{ID: "c2", PostID: "p1", ParentID: nil, Author: "A2", Text: "T2", CreatedAt: now},
					{ID: "c3", PostID: "p1", ParentID: nil, Author: "A3", Text: "T3", CreatedAt: now},
				},
				{Id: "p2", Limit: 5, Offset: 0}: {
					{ID: "c4", PostID: "p2", ParentID: nil, Author: "A4", Text: "T4", CreatedAt: now},
					{ID: "c5", PostID: "p2", ParentID: nil, Author: "A5", Text: "T5", CreatedAt: now},
				},
			},
			expectedErr: "",
		},
		{
			name: "success offset within range returning multiple values",
			keys: []entity.ParamKey{{Id: "p1", Limit: 5, Offset: 2}},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"req_id", "limit_val", "offset_val", "id", "post_id", "parent_id", "author", "text", "created_at"}).
					AddRow("p1", 5, 2, "c3", "p1", nil, "A3", "T3", now).
					AddRow("p1", 5, 2, "c4", "p1", nil, "A4", "T4", now).
					AddRow("p1", 5, 2, "c5", "p1", nil, "A5", "T5", now)
				mock.ExpectQuery("^SELECT.*FROM unnest").
					WithArgs([]string{"p1"}, []int{5}, []int{2}).WillReturnRows(rows)
			},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: "p1", Limit: 5, Offset: 2}: {
					{ID: "c3", PostID: "p1", ParentID: nil, Author: "A3", Text: "T3", CreatedAt: now},
					{ID: "c4", PostID: "p1", ParentID: nil, Author: "A4", Text: "T4", CreatedAt: now},
					{ID: "c5", PostID: "p1", ParentID: nil, Author: "A5", Text: "T5", CreatedAt: now},
				},
			},
			expectedErr: "",
		},
		{
			name: "success offset exceeds total",
			keys: []entity.ParamKey{{Id: "p1", Limit: 10, Offset: 1000}},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"req_id", "limit_val", "offset_val", "id", "post_id", "parent_id", "author", "text", "created_at"})
				mock.ExpectQuery("^SELECT.*FROM unnest").
					WithArgs([]string{"p1"}, []int{10}, []int{1000}).WillReturnRows(rows)
			},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: "p1", Limit: 10, Offset: 1000}: {},
			},
			expectedErr: "",
		},
		{
			name: "db error negative offset",
			keys: []entity.ParamKey{{Id: "p1", Limit: 10, Offset: -1}},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("^SELECT.*FROM unnest").
					WithArgs([]string{"p1"}, []int{10}, []int{-1}).WillReturnError(errors.New("OFFSET must not be negative"))
			},
			expectedRes: nil,
			expectedErr: "OFFSET must not be negative",
		},
		{
			name: "success limit zero",
			keys: []entity.ParamKey{{Id: "p1", Limit: 0, Offset: 0}},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"req_id", "limit_val", "offset_val", "id", "post_id", "parent_id", "author", "text", "created_at"})
				mock.ExpectQuery("^SELECT.*FROM unnest").
					WithArgs([]string{"p1"}, []int{0}, []int{0}).WillReturnRows(rows)
			},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: "p1", Limit: 0, Offset: 0}: {},
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, repo := setupTest(t)
			defer mock.Close()

			tt.mockSetup(mock)
			res, err := repo.GetCommentsByPostIDs(context.Background(), tt.keys)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestRepo_GetRepliesByParentIDs(t *testing.T) {
	now := time.Now()
	parentID1 := "c1"
	parentID2 := "c2"

	tests := []struct {
		name        string
		keys        []entity.ParamKey
		mockSetup   func(mock pgxmock.PgxPoolIface)
		expectedRes map[entity.ParamKey][]*entity.Comment
		expectedErr string
	}{
		{
			name: "success valid pagination with multiple parents and replies",
			keys: []entity.ParamKey{
				{Id: parentID1, Limit: 10, Offset: 0},
				{Id: parentID2, Limit: 5, Offset: 0},
			},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"req_id", "limit_val", "offset_val", "id", "post_id", "parent_id", "author", "text", "created_at"}).
					AddRow(parentID1, 10, 0, "r1", "p1", &parentID1, "A1", "T1", now).
					AddRow(parentID1, 10, 0, "r2", "p1", &parentID1, "A2", "T2", now).
					AddRow(parentID1, 10, 0, "r3", "p1", &parentID1, "A3", "T3", now).
					AddRow(parentID2, 5, 0, "r4", "p2", &parentID2, "A4", "T4", now).
					AddRow(parentID2, 5, 0, "r5", "p2", &parentID2, "A5", "T5", now)
				mock.ExpectQuery("^SELECT.*FROM unnest").
					WithArgs([]string{parentID1, parentID2}, []int{10, 5}, []int{0, 0}).WillReturnRows(rows)
			},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: parentID1, Limit: 10, Offset: 0}: {
					{ID: "r1", PostID: "p1", ParentID: &parentID1, Author: "A1", Text: "T1", CreatedAt: now},
					{ID: "r2", PostID: "p1", ParentID: &parentID1, Author: "A2", Text: "T2", CreatedAt: now},
					{ID: "r3", PostID: "p1", ParentID: &parentID1, Author: "A3", Text: "T3", CreatedAt: now},
				},
				{Id: parentID2, Limit: 5, Offset: 0}: {
					{ID: "r4", PostID: "p2", ParentID: &parentID2, Author: "A4", Text: "T4", CreatedAt: now},
					{ID: "r5", PostID: "p2", ParentID: &parentID2, Author: "A5", Text: "T5", CreatedAt: now},
				},
			},
			expectedErr: "",
		},
		{
			name: "success offset within range returning multiple values",
			keys: []entity.ParamKey{{Id: parentID1, Limit: 5, Offset: 2}},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"req_id", "limit_val", "offset_val", "id", "post_id", "parent_id", "author", "text", "created_at"}).
					AddRow(parentID1, 5, 2, "r3", "p1", &parentID1, "A3", "T3", now).
					AddRow(parentID1, 5, 2, "r4", "p1", &parentID1, "A4", "T4", now).
					AddRow(parentID1, 5, 2, "r5", "p1", &parentID1, "A5", "T5", now)
				mock.ExpectQuery("^SELECT.*FROM unnest").
					WithArgs([]string{parentID1}, []int{5}, []int{2}).WillReturnRows(rows)
			},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: parentID1, Limit: 5, Offset: 2}: {
					{ID: "r3", PostID: "p1", ParentID: &parentID1, Author: "A3", Text: "T3", CreatedAt: now},
					{ID: "r4", PostID: "p1", ParentID: &parentID1, Author: "A4", Text: "T4", CreatedAt: now},
					{ID: "r5", PostID: "p1", ParentID: &parentID1, Author: "A5", Text: "T5", CreatedAt: now},
				},
			},
			expectedErr: "",
		},
		{
			name: "success offset exceeds total",
			keys: []entity.ParamKey{{Id: parentID1, Limit: 5, Offset: 1000}},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"req_id", "limit_val", "offset_val", "id", "post_id", "parent_id", "author", "text", "created_at"})
				mock.ExpectQuery("^SELECT.*FROM unnest").
					WithArgs([]string{parentID1}, []int{5}, []int{1000}).WillReturnRows(rows)
			},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: parentID1, Limit: 5, Offset: 1000}: {},
			},
			expectedErr: "",
		},
		{
			name: "db error negative offset",
			keys: []entity.ParamKey{{Id: parentID1, Limit: 5, Offset: -1}},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				mock.ExpectQuery("^SELECT.*FROM unnest").
					WithArgs([]string{parentID1}, []int{5}, []int{-1}).WillReturnError(errors.New("OFFSET must not be negative"))
			},
			expectedRes: nil,
			expectedErr: "OFFSET must not be negative",
		},
		{
			name: "success limit zero",
			keys: []entity.ParamKey{{Id: parentID1, Limit: 0, Offset: 0}},
			mockSetup: func(mock pgxmock.PgxPoolIface) {
				rows := pgxmock.NewRows([]string{"req_id", "limit_val", "offset_val", "id", "post_id", "parent_id", "author", "text", "created_at"})
				mock.ExpectQuery("^SELECT.*FROM unnest").
					WithArgs([]string{parentID1}, []int{0}, []int{0}).WillReturnRows(rows)
			},
			expectedRes: map[entity.ParamKey][]*entity.Comment{
				{Id: parentID1, Limit: 0, Offset: 0}: {},
			},
			expectedErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock, repo := setupTest(t)
			defer mock.Close()

			tt.mockSetup(mock)
			res, err := repo.GetRepliesByParentIDs(context.Background(), tt.keys)

			if tt.expectedErr != "" {
				assert.ErrorContains(t, err, tt.expectedErr)
				assert.Nil(t, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedRes, res)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
