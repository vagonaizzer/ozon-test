// Package memory реализует хранилище данных в оперативной памяти.
package memory

import (
	"context"
	"sync"

	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
	post_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/post"
	"github.com/vagonaizer/ozon-test-assignment/pkg/customerror"
)

var _ post_iface.PostRepository = (*PostRepository)(nil)

type PostRepository struct {
	mu    sync.RWMutex
	store map[string]*post_entity.Post
	order []string
}

func NewPostRepository() *PostRepository {
	return &PostRepository{
		store: make(map[string]*post_entity.Post),
	}
}

func (r *PostRepository) Create(_ context.Context, p *post_entity.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := p.ID().String()
	r.store[id] = p
	r.order = append(r.order, id)
	return nil
}

func (r *PostRepository) GetByID(_ context.Context, id post_entity.PostID) (*post_entity.Post, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.store[id.String()]
	if !ok {
		return nil, customerror.NotFound("post", id.String())
	}
	return p, nil
}

func (r *PostRepository) List(_ context.Context, limit int, cursor string) ([]*post_entity.Post, string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	start := 0
	if cursor != "" {
		for i, id := range r.order {
			if id == cursor {
				start = i + 1
				break
			}
		}
	}

	end := start + limit
	if end > len(r.order) {
		end = len(r.order)
	}

	slice := r.order[start:end]
	result := make([]*post_entity.Post, 0, len(slice))
	for _, id := range slice {
		result = append(result, r.store[id])
	}

	var nextCursor string
	if len(result) > 0 {
		nextCursor = result[len(result)-1].ID().String()
	}

	return result, nextCursor, nil
}

func (r *PostRepository) Update(_ context.Context, p *post_entity.Post) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := p.ID().String()
	if _, ok := r.store[id]; !ok {
		return customerror.NotFound("post", id)
	}
	r.store[id] = p
	return nil
}
