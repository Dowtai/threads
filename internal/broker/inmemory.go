package broker

import (
	"context"
	"sync"
	"threads/internal/entity"
)

type CommentBroker interface {
	Subscribe(ctx context.Context, postID string) (chan *entity.Comment, error)
	Unsubscribe(ctx context.Context, postID string, ch chan *entity.Comment) error
	Publish(ctx context.Context, postID string, comment *entity.Comment) error
}

type inMemoryCommentBroker struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan *entity.Comment]struct{}
}

func NewInMemoryCommentBroker() *inMemoryCommentBroker {
	return &inMemoryCommentBroker{
		mu:          sync.RWMutex{},
		subscribers: make(map[string]map[chan *entity.Comment]struct{}),
	}
}

func (b *inMemoryCommentBroker) Subscribe(_ context.Context, postID string) (chan *entity.Comment, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	ch := make(chan *entity.Comment, 1)

	if b.subscribers[postID] == nil {
		b.subscribers[postID] = make(map[chan *entity.Comment]struct{})
	}

	b.subscribers[postID][ch] = struct{}{}

	return ch, nil
}

func (b *inMemoryCommentBroker) Unsubscribe(_ context.Context, postID string, ch chan *entity.Comment) error {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subs, ok := b.subscribers[postID]; ok {
		delete(subs, ch)
		close(ch)

		if len(subs) == 0 {
			delete(b.subscribers, postID)
		}
	}

	return nil
}

func (b *inMemoryCommentBroker) Publish(_ context.Context, postID string, comment *entity.Comment) error {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if subs, ok := b.subscribers[postID]; ok {
		for ch := range subs {
			select {
			case ch <- comment:
			default:
			}
		}
	}

	return nil
}
