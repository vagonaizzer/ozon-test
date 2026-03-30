// Package post реализует use-case логику для постов.
package post

import (
	"context"
	"fmt"

	"github.com/vagonaizer/ozon-test-assignment/internal/common"
	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
	post_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/post"
	"github.com/vagonaizer/ozon-test-assignment/pkg/customerror"
	"github.com/vagonaizer/ozon-test-assignment/pkg/pagination"
)

var _ post_iface.PostService = (*Service)(nil)

type Service struct {
	repo post_iface.PostRepository
}

func NewService(repo post_iface.PostRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, authorID, title, content string) (*post_entity.Post, error) {
	if authorID == "" {
		return nil, customerror.Validation("authorId", "cannot be empty")
	}

	author, err := post_entity.NewAuthorID(authorID)
	if err != nil {
		return nil, customerror.Validation("authorId", fmt.Sprintf("invalid ULID: %v", err))
	}

	titleVO, err := post_entity.NewPostTitle(title)
	if err != nil {
		return nil, customerror.Validation("title", err.Error())
	}

	contentVO, err := post_entity.NewPostContent(content)
	if err != nil {
		return nil, customerror.Validation("content", err.Error())
	}

	post := post_entity.NewPost(author, titleVO, contentVO)

	if err := s.repo.Create(ctx, post); err != nil {
		return nil, fmt.Errorf("create post: %w", err)
	}

	return post, nil
}

func (s *Service) GetByID(ctx context.Context, id post_entity.PostID) (*post_entity.Post, error) {
	post, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return post, nil
}

func (s *Service) List(ctx context.Context, limit int, cursor string) (*common.Page[*post_entity.Post], error) {
	limit = pagination.NormalizeLimit(limit)

	posts, nextCursor, err := s.repo.List(ctx, limit+1, pagination.DecodeCursor(cursor))
	if err != nil {
		return nil, fmt.Errorf("list posts: %w", err)
	}

	hasMore := len(posts) > limit
	if hasMore {
		posts = posts[:limit]
		nextCursor = pagination.EncodeCursor(posts[limit-1].ID().String())
	} else {
		nextCursor = ""
	}

	return &common.Page[*post_entity.Post]{
		Items:      posts,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *Service) ToggleComments(
	ctx context.Context,
	postID post_entity.PostID,
	authorID string,
	enabled bool,
) (*post_entity.Post, error) {
	post, err := s.repo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	if post.AuthorID().String() != authorID {
		return nil, customerror.Forbidden("only the author can toggle comments")
	}

	post.ToggleComments(enabled)

	if err := s.repo.Update(ctx, post); err != nil {
		return nil, fmt.Errorf("update post: %w", err)
	}

	return post, nil
}
