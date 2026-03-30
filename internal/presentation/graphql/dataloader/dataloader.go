// Package dataloader реализует батчевую загрузку данных для предотвращения N+1 запросов.
//
// Принцип работы:
//  1. За одну итерацию event-loop собираются все запросы к одному ресурсу.
//  2. Один батч-запрос отправляется в репозиторий.
//  3. Результаты раздаются всем ожидающим горутинам.
//
// Используется при рендеринге списка постов: вместо N запросов за
// комментариями делается один BatchGetByPostIDs.
package dataloader

import (
	"context"
	"sync"
	"time"

	comment_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/comment"
	post_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/post"
	comment_iface "github.com/vagonaizer/ozon-test-assignment/internal/domain/interfaces/comment"
)

type ctxKey struct{}

type Loaders struct {
	CommentsByPostID *CommentLoader
}

type CommentLoader struct {
	repo comment_iface.CommentRepository

	mu      sync.Mutex
	batch   []post_entity.PostID
	waiters []chan result
	timer   *time.Timer
	wait    time.Duration
}

type result struct {
	comments []*comment_entity.Comment
	err      error
}

func NewLoaders(repo comment_iface.CommentRepository) *Loaders {
	return &Loaders{
		CommentsByPostID: &CommentLoader{
			repo: repo,
			wait: 2 * time.Millisecond,
		},
	}
}

func (l *CommentLoader) Load(ctx context.Context, postID post_entity.PostID) ([]*comment_entity.Comment, error) {
	ch := make(chan result, 1)

	l.mu.Lock()
	l.batch = append(l.batch, postID)
	l.waiters = append(l.waiters, ch)

	if l.timer == nil {
		l.timer = time.AfterFunc(l.wait, func() { l.dispatch(ctx) })
	}
	l.mu.Unlock()

	r := <-ch
	return r.comments, r.err
}

func (l *CommentLoader) dispatch(ctx context.Context) {
	l.mu.Lock()
	keys := l.batch
	waiters := l.waiters
	l.batch = nil
	l.waiters = nil
	l.timer = nil
	l.mu.Unlock()

	data, err := l.repo.BatchGetByPostIDs(ctx, keys)

	for i, postID := range keys {
		if err != nil {
			waiters[i] <- result{err: err}
			continue
		}
		waiters[i] <- result{comments: data[postID]}
	}
}

func Attach(ctx context.Context, loaders *Loaders) context.Context {
	return context.WithValue(ctx, ctxKey{}, loaders)
}

func For(ctx context.Context) *Loaders {
	return ctx.Value(ctxKey{}).(*Loaders)
}
