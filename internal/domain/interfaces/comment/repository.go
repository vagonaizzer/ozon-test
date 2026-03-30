package comment

import (
	"context"

	comment_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/comment"
	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
)

type CommentRepository interface {
	Create(ctx context.Context, c *comment_entity.Comment) error
	GetByID(ctx context.Context, id comment_entity.CommentID) (*comment_entity.Comment, error)
	ListByPostID(ctx context.Context, postID post_entity.PostID, limit int, cursor string) ([]*comment_entity.Comment, string, error)
	ListByParentID(ctx context.Context, parentID comment_entity.CommentID, limit int, cursor string) ([]*comment_entity.Comment, string, error)
	BatchGetByPostIDs(ctx context.Context, postIDs []post_entity.PostID) (map[post_entity.PostID][]*comment_entity.Comment, error)
}
