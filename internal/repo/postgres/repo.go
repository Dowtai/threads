package postgres

import (
	"context"
	"errors"
	"threads/internal/entity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type repoImpl struct {
	pool *pgxpool.Pool
}

func New(pool *pgxpool.Pool) *repoImpl {
	return &repoImpl{
		pool: pool,
	}
}

func (r *repoImpl) GetPost(ctx context.Context, id string) (*entity.Post, error) {
	post := &entity.Post{}

	query := `SELECT id, author, title, content, comments_allowed FROM posts where id = $1;`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.Author,
		&post.Title,
		&post.Content,
		&post.CommentsAllowed,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("post not found")
		}

		return nil, err
	}

	return post, nil
}

func (r *repoImpl) GetComment(ctx context.Context, id string) (*entity.Comment, error) {
	comment := &entity.Comment{}

	query := `SELECT id, post_id, parent_id, author, text FROM comments where id = $1;`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.ParentID,
		&comment.Author,
		&comment.Text,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("comment not found")
		}

		return nil, err
	}

	return comment, nil
}

func (r *repoImpl) GetPosts(ctx context.Context, limit, offset int) ([]*entity.Post, error) {
	query := `SELECT id, author, title, content, comments_allowed FROM posts LIMIT $1 OFFSET $2;`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]*entity.Post, 0)

	for rows.Next() {
		post := &entity.Post{}

		err := rows.Scan(
			&post.ID,
			&post.Author,
			&post.Title,
			&post.Content,
			&post.CommentsAllowed,
		)
		if err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *repoImpl) GetPostComments(ctx context.Context, id string, limit, offset int) ([]*entity.Comment, error) {
	query := `SELECT id, post_id, parent_id, author, text FROM comments where post_id = $1 LIMIT $2 OFFSET $3;`

	rows, err := r.pool.Query(ctx, query, id, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*entity.Comment, 0)

	for rows.Next() {
		comment := &entity.Comment{}

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Author,
			&comment.Text,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *repoImpl) GetCommentReplies(ctx context.Context, id string, limit, offset int) ([]*entity.Comment, error) {
	query := `SELECT id, post_id, parent_id, author, text FROM comments where parent_id = $1 LIMIT $2 OFFSET $3;`

	rows, err := r.pool.Query(ctx, query, id, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := make([]*entity.Comment, 0)

	for rows.Next() {
		comment := &entity.Comment{}

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Author,
			&comment.Text,
		)
		if err != nil {
			return nil, err
		}
		comments = append(comments, comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (r *repoImpl) StorePost(ctx context.Context, post *entity.Post) (*entity.Post, error) {
	query := `INSERT INTO posts (id, author, title, content, comments_allowed) VALUES ($1, $2, $3, $4, $5);`

	_, err := r.pool.Exec(ctx, query,
		post.ID,
		post.Author,
		post.Title,
		post.Content,
		post.CommentsAllowed,
	)

	if err != nil {
		return nil, err
	}

	return post, nil
}

func (r *repoImpl) StoreComment(ctx context.Context, comment *entity.Comment) (*entity.Comment, error) {
	query := `INSERT INTO comments (id, post_id, parent_id, author, text) VALUES ($1, $2, $3, $4, $5);`

	_, err := r.pool.Exec(ctx, query,
		comment.ID,
		comment.PostID,
		comment.ParentID,
		comment.Author,
		comment.Text,
	)

	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *repoImpl) UpdateCommentsAllowed(ctx context.Context, postID string, allowed bool) (*entity.Post, error) {
	post := &entity.Post{}

	query := `UPDATE posts SET comments_allowed = $1 where id = $2 RETURNING id, author, title, content, comments_allowed;`

	err := r.pool.QueryRow(ctx, query, allowed, postID).Scan(
		&post.ID,
		&post.Author,
		&post.Title,
		&post.Content,
		&post.CommentsAllowed,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("post not found")
		}

		return nil, err
	}

	return post, nil
}

func (r *repoImpl) GetCommentsByPostIDs(ctx context.Context, ids []string) (map[string][]*entity.Comment, error) {
	query := `SELECT id, post_id, parent_id, author, text FROM comments WHERE post_id = ANY($1) AND parent_id IS NULL`

	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]*entity.Comment)
	for rows.Next() {
		comment := &entity.Comment{}

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Author,
			&comment.Text,
		)
		if err != nil {
			return nil, err
		}

		result[comment.PostID] = append(result[comment.PostID], comment)
	}

	return result, nil
}

func (r *repoImpl) GetRepliesByParentIDs(ctx context.Context, ids []string) (map[string][]*entity.Comment, error) {
	query := `SELECT id, post_id, parent_id, author, text FROM comments WHERE parent_id = ANY($1)`

	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string][]*entity.Comment)
	for rows.Next() {
		comment := &entity.Comment{}

		err := rows.Scan(
			&comment.ID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Author,
			&comment.Text,
		)
		if err != nil {
			return nil, err
		}

		if comment.ParentID != nil {
			result[*comment.ParentID] = append(result[*comment.ParentID], comment)
		} else {
			return nil, errors.New("something terrible went wrong: parent not found but unreachable case")
		}
	}

	return result, nil
}
