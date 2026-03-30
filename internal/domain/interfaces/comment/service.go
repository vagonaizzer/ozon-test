package comment

import (
	"context"

	"github.com/vagonaizer/ozon-test-assignment/internal/common"
	comment_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/comment"
	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
)


type CommentService interface {
	Create(
		ctx context.Context,
		postID post_entity.PostID,
		authorID string,
		parentID *comment_entity.CommentID,
		text string,
	) (*comment_entity.Comment, error)

	GetByID(ctx context.Context, id comment_entity.CommentID) (*comment_entity.Comment, error)

	ListByPostID(
		ctx context.Context,
		postID post_entity.PostID,
		limit int,
		cursor string,
	) (*common.Page[*comment_entity.Comment], error)

	ListByParentID(
		ctx context.Context,
		parentID comment_entity.CommentID,
		limit int,
		cursor string,
	) (*common.Page[*comment_entity.Comment], error)
}
