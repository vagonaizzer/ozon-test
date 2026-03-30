// Package postgres реализует репозитории на базе PostgreSQL через pgx/v5.
package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
	post_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/post"
	"github.com/vagonaizer/ozon-test-assignment/pkg/customerror"
)

// Проверка интерфейса на этапе компиляции.
var _ post_iface.PostRepository = (*PostRepository)(nil)

type PostRepository struct {
	pool *pgxpool.Pool
}

func NewPostRepository(pool *pgxpool.Pool) *PostRepository {
	return &PostRepository{pool: pool}
}

func (r *PostRepository) Create(ctx context.Context, p *post_entity.Post) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO posts (id, author_id, title, content, comments_enabled, created_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		p.ID().String(),
		p.AuthorID().String(),
		p.Title().String(),
		p.Content().String(),
		p.CommentsEnabled(),
		p.CreatedAt(),
	)
	if err != nil {
		return fmt.Errorf("insert post: %w", err)
	}
	return nil
}

func (r *PostRepository) GetByID(ctx context.Context, id post_entity.PostID) (*post_entity.Post, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, author_id, title, content, comments_enabled, created_at
		 FROM posts WHERE id = $1`,
		id.String(),
	)

	p, err := scanPost(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customerror.NotFound("post", id.String())
		}
		return nil, fmt.Errorf("get post by id: %w", err)
	}
	return p, nil
}

// List использует курсорную пагинацию по ULID (лексикографически сортируемому).
func (r *PostRepository) List(ctx context.Context, limit int, cursor string) ([]*post_entity.Post, string, error) {
	var (
		rows pgx.Rows
		err  error
	)

	if cursor == "" {
		rows, err = r.pool.Query(ctx,
			`SELECT id, author_id, title, content, comments_enabled, created_at
			 FROM posts ORDER BY id ASC LIMIT $1`,
			limit,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, author_id, title, content, comments_enabled, created_at
			 FROM posts WHERE id > $1 ORDER BY id ASC LIMIT $2`,
			cursor, limit,
		)
	}
	if err != nil {
		return nil, "", fmt.Errorf("list posts: %w", err)
	}
	defer rows.Close()

	posts, err := collectPosts(rows)
	if err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(posts) > 0 {
		nextCursor = posts[len(posts)-1].ID().String()
	}

	return posts, nextCursor, nil
}

func (r *PostRepository) Update(ctx context.Context, p *post_entity.Post) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE posts SET title=$1, content=$2, comments_enabled=$3 WHERE id=$4`,
		p.Title().String(),
		p.Content().String(),
		p.CommentsEnabled(),
		p.ID().String(),
	)
	if err != nil {
		return fmt.Errorf("update post: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return customerror.NotFound("post", p.ID().String())
	}
	return nil
}

func collectPosts(rows pgx.Rows) ([]*post_entity.Post, error) {
	var posts []*post_entity.Post
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

type scanner interface{ Scan(dest ...any) error }

func scanPost(s scanner) (*post_entity.Post, error) {
	var (
		idStr, authorIDStr string
		titleStr, contentStr string
		commentsEnabled    bool
		createdAt          time.Time
	)

	if err := s.Scan(&idStr, &authorIDStr, &titleStr, &contentStr, &commentsEnabled, &createdAt); err != nil {
		return nil, err
	}

	postID, err := post_entity.ParsePostID(idStr)
	if err != nil {
		return nil, fmt.Errorf("parse post id %q: %w", idStr, err)
	}

	authorID, err := post_entity.NewAuthorID(authorIDStr)
	if err != nil {
		return nil, fmt.Errorf("parse author id %q: %w", authorIDStr, err)
	}

	titleVO, err := post_entity.NewPostTitle(titleStr)
	if err != nil {
		return nil, fmt.Errorf("parse title: %w", err)
	}

	contentVO, err := post_entity.NewPostContent(contentStr)
	if err != nil {
		return nil, fmt.Errorf("parse content: %w", err)
	}

	p := post_entity.NewPost(authorID, titleVO, contentVO)
	post_entity.RestorePost(p, postID, authorID, commentsEnabled, createdAt)
	return p, nil
}
