package post

import (
	"context"

	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
)

type PostRepository interface {
	Create(ctx context.Context, p *post_entity.Post) error
	GetByID(ctx context.Context, id post_entity.PostID) (*post_entity.Post, error)
	List(ctx context.Context, limit int, cursor string) ([]*post_entity.Post, string, error)
	Update(ctx context.Context, p *post_entity.Post) error
}
