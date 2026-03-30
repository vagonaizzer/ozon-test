package subscription

import (
	"sync"

	comment_entity "github.com/vagonaizer/ozon-test-assignment/internal/domain/entity/comment"
)

const bufSize = 16

type Manager struct {
	mu          sync.RWMutex
	subscribers map[comment_entity.PostID][]chan *comment_entity.Comment
}

func New() *Manager {
	return &Manager{
		subscribers: make(map[comment_entity.PostID][]chan *comment_entity.Comment),
	}
}

func (m *Manager) Subscribe(postID comment_entity.PostID) (<-chan *comment_entity.Comment, func()) {
	ch := make(chan *comment_entity.Comment, bufSize)

	m.mu.Lock()
	m.subscribers[postID] = append(m.subscribers[postID], ch)
	m.mu.Unlock()

	unsubscribe := func() {
		m.mu.Lock()
		defer m.mu.Unlock()

		subs := m.subscribers[postID]
		for i, s := range subs {
			if s == ch {
				m.subscribers[postID] = append(subs[:i], subs[i+1:]...)
				break
			}
		}
		close(ch)
	}

	return ch, unsubscribe
}

func (m *Manager) Publish(postID comment_entity.PostID, c *comment_entity.Comment) {
	m.mu.RLock()
	subs := make([]chan *comment_entity.Comment, len(m.subscribers[postID]))
	copy(subs, m.subscribers[postID])
	m.mu.RUnlock()

	for _, ch := range subs {
		select {
		case ch <- c:
		default:
		}
	}
}
