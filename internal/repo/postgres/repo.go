package postgres

import (
	"context"
	"errors"
	"threads/internal/entity"

	"github.com/jackc/pgx/v5"
)

type DBPool interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type repoImpl struct {
	pool DBPool
}

func New(pool DBPool) *repoImpl {
	return &repoImpl{
		pool: pool,
	}
}

func (r *repoImpl) GetPost(ctx context.Context, id string) (*entity.Post, error) {
	post := &entity.Post{}

	query := `SELECT id, author, title, content, comments_allowed, created_at FROM posts where id = $1;`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&post.ID,
		&post.Author,
		&post.Title,
		&post.Content,
		&post.CommentsAllowed,
		&post.CreatedAt,
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

	query := `SELECT id, post_id, parent_id, author, text, created_at FROM comments where id = $1;`

	err := r.pool.QueryRow(ctx, query, id).Scan(
		&comment.ID,
		&comment.PostID,
		&comment.ParentID,
		&comment.Author,
		&comment.Text,
		&comment.CreatedAt,
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
	query := `SELECT id, author, title, content, comments_allowed, created_at FROM posts 
    		  ORDER BY created_at DESC LIMIT $1 OFFSET $2;`

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
			&post.CreatedAt,
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

func (r *repoImpl) StorePost(ctx context.Context, post *entity.Post) (*entity.Post, error) {
	query := `INSERT INTO posts (id, author, title, content, comments_allowed, created_at) 
							VALUES ($1, $2, $3, $4, $5, NOW() AT TIME ZONE 'utc')
							RETURNING created_at;`

	err := r.pool.QueryRow(ctx, query,
		post.ID,
		post.Author,
		post.Title,
		post.Content,
		post.CommentsAllowed,
	).Scan(&post.CreatedAt)

	if err != nil {
		return nil, err
	}

	return post, nil
}

func (r *repoImpl) StoreComment(ctx context.Context, comment *entity.Comment) (*entity.Comment, error) {
	query := `INSERT INTO comments (id, post_id, parent_id, author, text, created_at) 
							VALUES ($1, $2, $3, $4, $5, NOW() AT TIME ZONE 'utc')
							RETURNING created_at;`

	err := r.pool.QueryRow(ctx, query,
		comment.ID,
		comment.PostID,
		comment.ParentID,
		comment.Author,
		comment.Text,
	).Scan(&comment.CreatedAt)

	if err != nil {
		return nil, err
	}

	return comment, nil
}

func (r *repoImpl) UpdateCommentsAllowed(ctx context.Context, postID string, allowed bool) (*entity.Post, error) {
	post := &entity.Post{}

	query := `UPDATE posts SET comments_allowed = $1 where id = $2 RETURNING id, author, title, content, comments_allowed, created_at;`

	err := r.pool.QueryRow(ctx, query, allowed, postID).Scan(
		&post.ID,
		&post.Author,
		&post.Title,
		&post.Content,
		&post.CommentsAllowed,
		&post.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("post not found")
		}

		return nil, err
	}

	return post, nil
}

func (r *repoImpl) getCommentsByIDs(ctx context.Context, keys []entity.ParamKey, query string) (map[entity.ParamKey][]*entity.Comment, error) {
	ids := make([]string, len(keys))
	limits := make([]int, len(keys))
	offsets := make([]int, len(keys))

	for i, key := range keys {
		ids[i] = key.Id
		limits[i] = key.Limit
		offsets[i] = key.Offset
	}

	rows, err := r.pool.Query(ctx, query, ids, limits, offsets)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[entity.ParamKey][]*entity.Comment)
	for rows.Next() {
		var reqID string
		var reqLimit int
		var reqOffset int
		comment := &entity.Comment{}

		err := rows.Scan(
			&reqID,
			&reqLimit,
			&reqOffset,
			&comment.ID,
			&comment.PostID,
			&comment.ParentID,
			&comment.Author,
			&comment.Text,
			&comment.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		key := entity.ParamKey{
			Id:     reqID,
			Limit:  reqLimit,
			Offset: reqOffset,
		}

		result[key] = append(result[key], comment)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	for _, key := range keys {
		if _, ok := result[key]; !ok {
			result[key] = make([]*entity.Comment, 0)
		}
	}

	return result, nil
}

func (r *repoImpl) GetCommentsByPostIDs(ctx context.Context, keys []entity.ParamKey) (map[entity.ParamKey][]*entity.Comment, error) {
	query := `SELECT
					req.id, req.limit_val, req.offset_val,
                    c.id, c.post_id, c.parent_id, c.author, c.text, c.created_at
              FROM unnest($1::uuid[], $2::int[], $3::int[]) AS req(id, limit_val, offset_val)
              CROSS JOIN LATERAL(
                  SELECT id, post_id, parent_id, author, text, created_at
                  FROM comments
                  WHERE post_id = req.id AND parent_id IS NULL
                  ORDER BY created_at DESC
                  LIMIT req.limit_val OFFSET req.offset_val
              ) c`

	return r.getCommentsByIDs(ctx, keys, query)
}

func (r *repoImpl) GetRepliesByParentIDs(ctx context.Context, keys []entity.ParamKey) (map[entity.ParamKey][]*entity.Comment, error) {
	query := `SELECT
					req.id, req.limit_val, req.offset_val,
                    c.id, c.post_id, c.parent_id, c.author, c.text, c.created_at
              FROM unnest($1::uuid[], $2::int[], $3::int[]) AS req(id, limit_val, offset_val)
              CROSS JOIN LATERAL(
                  SELECT id, post_id, parent_id, author, text, created_at
                  FROM comments
                  WHERE parent_id = req.id
                  ORDER BY created_at DESC
                  LIMIT req.limit_val OFFSET req.offset_val
              ) c`

	return r.getCommentsByIDs(ctx, keys, query)
}
