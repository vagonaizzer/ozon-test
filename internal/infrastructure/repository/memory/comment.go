package memory

import (
	"context"
	"sync"

	comment_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/comment"
	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
	comment_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/comment"
	"github.com/vagonaizer/ozon-test-assignment/pkg/customerror"
	"github.com/vagonaizer/ozon-test-assignment/pkg/pagination"
)

var _ comment_iface.CommentRepository = (*CommentRepository)(nil)

// Структура:
//   - store: все комментарии по ID
//   - byPost:   postID  -> []commentID (только корневые)
//   - byParent: parentID -> []commentID (дочерние)
//
// O(1) доступ по ID и O(n) по посту/родителю.
type CommentRepository struct {
	mu       sync.RWMutex
	store    map[comment_entity.CommentID]*comment_entity.Comment
	byPost   map[comment_entity.PostID][]comment_entity.CommentID
	byParent map[comment_entity.CommentID][]comment_entity.CommentID
}

func NewCommentRepository() *CommentRepository {
	return &CommentRepository{
		store:    make(map[comment_entity.CommentID]*comment_entity.Comment),
		byPost:   make(map[comment_entity.PostID][]comment_entity.CommentID),
		byParent: make(map[comment_entity.CommentID][]comment_entity.CommentID),
	}
}

func (r *CommentRepository) Create(_ context.Context, c *comment_entity.Comment) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[c.ID()] = c

	if c.IsRoot() {
		r.byPost[c.PostID()] = append(r.byPost[c.PostID()], c.ID())
	} else {
		pid := *c.ParentID()
		r.byParent[pid] = append(r.byParent[pid], c.ID())
	}
	return nil
}

func (r *CommentRepository) GetByID(_ context.Context, id comment_entity.CommentID) (*comment_entity.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.store[id]
	if !ok {
		return nil, customerror.NotFound("comment", pagination.EncodeIntCursor(int64(id)))
	}
	return c, nil
}

func (r *CommentRepository) ListByPostID(
	_ context.Context,
	postID post_entity.PostID,
	limit int,
	cursor string,
) ([]*comment_entity.Comment, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.byPost[comment_entity.PostID(postID.String())]
	return r.paginate(ids, limit, cursor)
}

func (r *CommentRepository) ListByParentID(
	_ context.Context,
	parentID comment_entity.CommentID,
	limit int,
	cursor string,
) ([]*comment_entity.Comment, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ids := r.byParent[parentID]
	return r.paginate(ids, limit, cursor)
}

func (r *CommentRepository) BatchGetByPostIDs(
	_ context.Context,
	postIDs []post_entity.PostID,
) (map[post_entity.PostID][]*comment_entity.Comment, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[post_entity.PostID][]*comment_entity.Comment, len(postIDs))
	for _, pid := range postIDs {
		key := comment_entity.PostID(pid.String())
		ids := r.byPost[key]
		comments := make([]*comment_entity.Comment, 0, len(ids))
		for _, id := range ids {
			if c, ok := r.store[id]; ok {
				comments = append(comments, c)
			}
		}
		result[pid] = comments
	}
	return result, nil
}

func (r *CommentRepository) paginate(
	ids []comment_entity.CommentID,
	limit int,
	cursor string,
) ([]*comment_entity.Comment, string, error) {
	afterID := comment_entity.CommentID(pagination.DecodeIntCursor(cursor))

	start := 0
	if afterID > 0 {
		for i, id := range ids {
			if id == afterID {
				start = i + 1
				break
			}
		}
	}

	end := start + limit
	if end > len(ids) {
		end = len(ids)
	}

	slice := ids[start:end]
	result := make([]*comment_entity.Comment, 0, len(slice))
	for _, id := range slice {
		if c, ok := r.store[id]; ok {
			result = append(result, c)
		}
	}

	var nextCursor string
	if len(result) > 0 {
		nextCursor = pagination.EncodeIntCursor(int64(result[len(result)-1].ID()))
	}

	return result, nextCursor, nil
}
