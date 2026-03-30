package resolver

import (
	comment_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/comment"
	post_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/post"
	"github.com/vagonaizer/ozon-test-assignment/internal/presentation/graphql/subscription"
)

type Resolver struct {
	PostService    post_iface.PostService
	CommentService comment_iface.CommentService
	SubManager     *subscription.Manager
}

func NewResolver(
	ps post_iface.PostService,
	cs comment_iface.CommentService,
	sm *subscription.Manager,
) *Resolver {
	return &Resolver{
		PostService:    ps,
		CommentService: cs,
		SubManager:     sm,
	}
}
