// Package comment реализует use-case логику для комментариев.
package comment

import (
	"context"
	"fmt"

	"github.com/vagonaizer/ozon-test-assignment/internal/common"
	comment_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/comment"
	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
	comment_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/comment"
	post_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/post"
	"github.com/vagonaizer/ozon-test-assignment/internal/presentation/graphql/subscription"
	"github.com/vagonaizer/ozon-test-assignment/pkg/customerror"
	"github.com/vagonaizer/ozon-test-assignment/pkg/pagination"
)

var _ comment_iface.CommentService = (*Service)(nil)

type Service struct {
	commentRepo comment_iface.CommentRepository
	postRepo    post_iface.PostRepository
	subManager  *subscription.Manager
}

func NewService(
	commentRepo comment_iface.CommentRepository,
	postRepo post_iface.PostRepository,
	subManager *subscription.Manager,
) *Service {
	return &Service{
		commentRepo: commentRepo,
		postRepo:    postRepo,
		subManager:  subManager,
	}
}

func (s *Service) Create(
	ctx context.Context,
	postID post_entity.PostID,
	authorID string,
	parentID *comment_entity.CommentID,
	text string,
) (*comment_entity.Comment, error) {
	post, err := s.postRepo.GetByID(ctx, postID)
	if err != nil {
		return nil, err
	}

	if !post.CommentsEnabled() {
		return nil, customerror.Forbidden("comments are disabled for this post")
	}

	if parentID != nil {
		if _, err := s.commentRepo.GetByID(ctx, *parentID); err != nil {
			return nil, fmt.Errorf("parent comment: %w", err)
		}
	}

	c, err := comment_entity.NewComment(
		comment_entity.PostID(postID.String()),
		comment_entity.AuthorID(authorID),
		parentID,
		comment_entity.CommentContent(text),
	)
	if err != nil {
		return nil, customerror.Validation("text", err.Error())
	}

	if err := s.commentRepo.Create(ctx, c); err != nil {
		return nil, fmt.Errorf("create comment: %w", err)
	}

	s.subManager.Publish(comment_entity.PostID(postID.String()), c)

	return c, nil
}

func (s *Service) GetByID(ctx context.Context, id comment_entity.CommentID) (*comment_entity.Comment, error) {
	return s.commentRepo.GetByID(ctx, id)
}

func (s *Service) ListByPostID(
	ctx context.Context,
	postID post_entity.PostID,
	limit int,
	cursor string,
) (*common.Page[*comment_entity.Comment], error) {
	limit = pagination.NormalizeLimit(limit)

	comments, _, err := s.commentRepo.ListByPostID(ctx, postID, limit+1, cursor)
	if err != nil {
		return nil, fmt.Errorf("list comments by post: %w", err)
	}

	return buildPage(comments, limit), nil
}

func (s *Service) ListByParentID(
	ctx context.Context,
	parentID comment_entity.CommentID,
	limit int,
	cursor string,
) (*common.Page[*comment_entity.Comment], error) {
	limit = pagination.NormalizeLimit(limit)

	comments, _, err := s.commentRepo.ListByParentID(ctx, parentID, limit+1, cursor)
	if err != nil {
		return nil, fmt.Errorf("list replies: %w", err)
	}

	return buildPage(comments, limit), nil
}

func buildPage(comments []*comment_entity.Comment, limit int) *common.Page[*comment_entity.Comment] {
	hasMore := len(comments) > limit
	if hasMore {
		comments = comments[:limit]
	}

	var nextCursor string
	if hasMore && len(comments) > 0 {
		nextCursor = pagination.EncodeIntCursor(int64(comments[len(comments)-1].ID()))
	}

	return &common.Page[*comment_entity.Comment]{
		Items:      comments,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}
}
