package post

import (
	"context"

	"github.com/vagonaizer/ozon-test-assignment/internal/common"
	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
)

type PostService interface {
	Create(ctx context.Context, authorID, title, content string) (*post_entity.Post, error)

	GetByID(ctx context.Context, id post_entity.PostID) (*post_entity.Post, error)

	List(ctx context.Context, limit int, cursor string) (*common.Page[*post_entity.Post], error)

	ToggleComments(ctx context.Context, postID post_entity.PostID, authorID string, enabled bool) (*post_entity.Post, error)
}
