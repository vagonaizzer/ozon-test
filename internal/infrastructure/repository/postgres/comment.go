package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	comment_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/comment"
	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
	comment_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/comment"
	"github.com/vagonaizer/ozon-test-assignment/pkg/customerror"
	"github.com/vagonaizer/ozon-test-assignment/pkg/pagination"
)

// Проверка интерфейса на этапе компиляции.
var _ comment_iface.CommentRepository = (*CommentRepository)(nil)

type CommentRepository struct {
	pool *pgxpool.Pool
}

func NewCommentRepository(pool *pgxpool.Pool) *CommentRepository {
	return &CommentRepository{pool: pool}
}

// Create вставляет комментарий и устанавливает DB-сгенерированный ID обратно в сущность.
func (r *CommentRepository) Create(ctx context.Context, c *comment_entity.Comment) error {
	var newID int64

	err := r.pool.QueryRow(ctx,
		`INSERT INTO comments (post_id, author_id, parent_id, content, created_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		string(c.PostID()),
		string(c.AuthorID()),
		nullableCommentID(c.ParentID()),
		string(c.Text()),
		c.CreatedAt(),
	).Scan(&newID)
	if err != nil {
		return fmt.Errorf("insert comment: %w", err)
	}

	c.SetID(comment_entity.CommentID(newID))
	return nil
}

func (r *CommentRepository) GetByID(ctx context.Context, id comment_entity.CommentID) (*comment_entity.Comment, error) {
	row := r.pool.QueryRow(ctx,
		`SELECT id, post_id, author_id, parent_id, content, created_at
		 FROM comments WHERE id = $1`,
		int64(id),
	)

	c, err := scanComment(row)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, customerror.NotFound("comment", pagination.EncodeIntCursor(int64(id)))
		}
		return nil, fmt.Errorf("get comment by id: %w", err)
	}
	return c, nil
}

// ListByPostID возвращает корневые комментарии поста (parent_id IS NULL) с пагинацией.
func (r *CommentRepository) ListByPostID(
	ctx context.Context,
	postID post_entity.PostID,
	limit int,
	cursor string,
) ([]*comment_entity.Comment, string, error) {
	afterID := pagination.DecodeIntCursor(cursor)

	var (
		rows pgx.Rows
		err  error
	)

	if afterID == 0 {
		rows, err = r.pool.Query(ctx,
			`SELECT id, post_id, author_id, parent_id, content, created_at
			 FROM comments
			 WHERE post_id = $1 AND parent_id IS NULL
			 ORDER BY id ASC LIMIT $2`,
			postID.String(), limit,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, post_id, author_id, parent_id, content, created_at
			 FROM comments
			 WHERE post_id = $1 AND parent_id IS NULL AND id > $2
			 ORDER BY id ASC LIMIT $3`,
			postID.String(), afterID, limit,
		)
	}
	if err != nil {
		return nil, "", fmt.Errorf("list comments by post: %w", err)
	}
	defer rows.Close()

	return collectComments(rows)
}

// ListByParentID возвращает дочерние комментарии с пагинацией.
func (r *CommentRepository) ListByParentID(
	ctx context.Context,
	parentID comment_entity.CommentID,
	limit int,
	cursor string,
) ([]*comment_entity.Comment, string, error) {
	afterID := pagination.DecodeIntCursor(cursor)

	var (
		rows pgx.Rows
		err  error
	)

	if afterID == 0 {
		rows, err = r.pool.Query(ctx,
			`SELECT id, post_id, author_id, parent_id, content, created_at
			 FROM comments
			 WHERE parent_id = $1
			 ORDER BY id ASC LIMIT $2`,
			int64(parentID), limit,
		)
	} else {
		rows, err = r.pool.Query(ctx,
			`SELECT id, post_id, author_id, parent_id, content, created_at
			 FROM comments
			 WHERE parent_id = $1 AND id > $2
			 ORDER BY id ASC LIMIT $3`,
			int64(parentID), afterID, limit,
		)
	}
	if err != nil {
		return nil, "", fmt.Errorf("list replies: %w", err)
	}
	defer rows.Close()

	return collectComments(rows)
}

// BatchGetByPostIDs загружает корневые комментарии для набора постов одним запросом.
// Предотвращает N+1 при рендеринге списка постов.
func (r *CommentRepository) BatchGetByPostIDs(
	ctx context.Context,
	postIDs []post_entity.PostID,
) (map[post_entity.PostID][]*comment_entity.Comment, error) {
	if len(postIDs) == 0 {
		return map[post_entity.PostID][]*comment_entity.Comment{}, nil
	}

	ids := make([]string, len(postIDs))
	for i, pid := range postIDs {
		ids[i] = pid.String()
	}

	rows, err := r.pool.Query(ctx,
		`SELECT id, post_id, author_id, parent_id, content, created_at
		 FROM comments
		 WHERE post_id = ANY($1) AND parent_id IS NULL
		 ORDER BY id ASC`,
		ids,
	)
	if err != nil {
		return nil, fmt.Errorf("batch get comments: %w", err)
	}
	defer rows.Close()

	result := make(map[post_entity.PostID][]*comment_entity.Comment, len(postIDs))
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, err
		}
		pid, err := post_entity.ParsePostID(string(c.PostID()))
		if err != nil {
			return nil, fmt.Errorf("parse post id in batch: %w", err)
		}
		result[pid] = append(result[pid], c)
	}
	return result, rows.Err()
}

func collectComments(rows pgx.Rows) ([]*comment_entity.Comment, string, error) {
	var comments []*comment_entity.Comment
	for rows.Next() {
		c, err := scanComment(rows)
		if err != nil {
			return nil, "", err
		}
		comments = append(comments, c)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}

	var nextCursor string
	if len(comments) > 0 {
		nextCursor = pagination.EncodeIntCursor(int64(comments[len(comments)-1].ID()))
	}
	return comments, nextCursor, nil
}

func scanComment(s scanner) (*comment_entity.Comment, error) {
	var (
		id                             int64
		postIDStr, authorIDStr, content string
		parentID                       *int64
		createdAt                      time.Time
	)

	if err := s.Scan(&id, &postIDStr, &authorIDStr, &parentID, &content, &createdAt); err != nil {
		return nil, err
	}

	var parentCommentID *comment_entity.CommentID
	if parentID != nil {
		pid := comment_entity.CommentID(*parentID)
		parentCommentID = &pid
	}

	c, err := comment_entity.NewComment(
		comment_entity.PostID(postIDStr),
		comment_entity.AuthorID(authorIDStr),
		parentCommentID,
		comment_entity.CommentContent(content),
	)
	if err != nil {
		return nil, fmt.Errorf("restore comment: %w", err)
	}

	c.SetID(comment_entity.CommentID(id))
	return c, nil
}

// nullableCommentID конвертирует *CommentID в *int64 для pgx.
func nullableCommentID(id *comment_entity.CommentID) *int64 {
	if id == nil {
		return nil
	}
	v := int64(*id)
	return &v
}
