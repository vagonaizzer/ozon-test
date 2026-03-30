package resolver

import (
	"fmt"

	"github.com/vagonaizer/ozon-test-assignment/api/graphql/generated"
	comment_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/comment"
	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
)

func postToGQL(p *post_entity.Post) *generated.Post {
	return &generated.Post{
		ID:              p.ID().String(),
		AuthorID:        p.AuthorID().String(),
		Title:           p.Title().String(),
		Content:         p.Content().String(),
		CommentsEnabled: p.CommentsEnabled(),
		CreatedAt:       p.CreatedAt(),
	}
}

func commentToGQL(c *comment_entity.Comment) *generated.Comment {
	out := &generated.Comment{
		ID:        fmt.Sprintf("%d", c.ID()),
		PostID:    string(c.PostID()),
		AuthorID:  string(c.AuthorID()),
		Text:      string(c.Text()),
		CreatedAt: c.CreatedAt(),
	}
	if pid := c.ParentID(); pid != nil {
		s := fmt.Sprintf("%d", *pid)
		out.ParentID = &s
	}
	return out
}

func commentPageToGQL(comments []*comment_entity.Comment, nextCursor string, hasMore bool) *generated.CommentPage {
	gqlComments := make([]*generated.Comment, len(comments))
	for i, c := range comments {
		gqlComments[i] = commentToGQL(c)
	}

	var nc *string
	if nextCursor != "" {
		nc = &nextCursor
	}

	return &generated.CommentPage{
		Comments:   gqlComments,
		NextCursor: nc,
		HasMore:    hasMore,
	}
}
