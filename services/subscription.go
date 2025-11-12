package services

import (
	"fmt"
	"sync"

	"github.com/iden3/notification-service/log"
	"github.com/pkg/errors"
)

var (
	// ErrMaxSubscriptionsReached is returned when a user has reached the maximum number of subscriptions
	ErrMaxSubscriptionsReached = errors.New("maximum number of subscriptions reached")
)

type Subscriber struct {
	userDID string
}

func NewSubscriber(userDID string) Subscriber {
	return Subscriber{
		userDID: userDID,
	}
}

type SubscriptionService struct {
	lock        sync.RWMutex
	subscribers map[Subscriber][]chan NotificationPayload

	maxSubscriptionsPerUser int
}

func NewSubscriptionService(maxSubscriptionsPerUser int) *SubscriptionService {
	return &SubscriptionService{
		lock:        sync.RWMutex{},
		subscribers: make(map[Subscriber][]chan NotificationPayload),

		maxSubscriptionsPerUser: maxSubscriptionsPerUser,
	}
}

func (s *SubscriptionService) Subscribe(userDID string) (<-chan NotificationPayload, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	subscriber := NewSubscriber(userDID)
	// Check if max subscriptions reached
	// if maxSubscriptionsPerUser is 0, then unlimited subscriptions are allowed
	if s.maxSubscriptionsPerUser > 0 &&
		len(s.subscribers[subscriber]) > s.maxSubscriptionsPerUser {
		return nil, fmt.Errorf("%w: allowed connectios: %v",
			ErrMaxSubscriptionsReached, s.maxSubscriptionsPerUser)
	}

	ch := make(chan NotificationPayload)
	s.subscribers[subscriber] = append(s.subscribers[subscriber], ch)
	return ch, nil
}

func (s *SubscriptionService) Unsubscribe(userDID string, uch <-chan NotificationPayload) {
	s.lock.Lock()
	defer s.lock.Unlock()

	subscriber := NewSubscriber(userDID)
	channels, ok := s.subscribers[subscriber]
	if !ok {
		return
	}

	for id, c := range channels {
		if c == uch {
			close(c)
			s.subscribers[subscriber] = append(channels[:id], channels[id+1:]...)
			break
		}
	}

	// if no more channels, remove subscriber entry
	if len(s.subscribers[subscriber]) == 0 {
		delete(s.subscribers, subscriber)
	}
}

func (s *SubscriptionService) Notify(userDID string, payload NotificationPayload) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if channels, exists := s.subscribers[NewSubscriber(userDID)]; exists {
		for _, c := range channels {
			select {
			case c <- payload:
			default:
				log.Warnf("Notification dropped for user %s: channel is full", userDID)
			}
		}
	}
}
