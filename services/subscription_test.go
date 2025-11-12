package services

import (
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNotify_SingleSubscriber(t *testing.T) {
	service := NewSubscriptionService(10)
	userDID := "did:example:123"
	payload := NotificationPayload{ID: "1"}

	ch, err := service.Subscribe(userDID)
	require.NoError(t, err)

	go service.Notify(userDID, payload)

	select {
	case received := <-ch:
		require.Equal(t, payload, received)
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for notification")
	}
}

func TestNotify_MultipleSubscribersForSameUser(t *testing.T) {
	service := NewSubscriptionService(10)
	userDID := "did:example:123"
	payload := NotificationPayload{ID: "2"}

	ch1, err := service.Subscribe(userDID)
	require.NoError(t, err)
	ch2, err := service.Subscribe(userDID)
	require.NoError(t, err)
	ch3, err := service.Subscribe(userDID)
	require.NoError(t, err)

	var wg sync.WaitGroup
	wg.Add(3)

	checkChannel := func(ch <-chan NotificationPayload) {
		defer wg.Done()
		select {
		case received := <-ch:
			require.Equal(t, payload, received)
		case <-time.After(time.Second):
			t.Error("timeout waiting for notification")
		}
	}

	go checkChannel(ch1)
	go checkChannel(ch2)
	go checkChannel(ch3)

	// Give goroutines time to start waiting on channels
	time.Sleep(50 * time.Millisecond)

	service.Notify(userDID, payload)
	wg.Wait()
}

func TestNotify_NoSubscribers(t *testing.T) {
	service := NewSubscriptionService(10)
	userDID := "did:example:nonexistent"
	payload := NotificationPayload{ID: "3"}

	// Should not panic or block
	service.Notify(userDID, payload)
}

func TestNotify_DifferentUsers(t *testing.T) {
	service := NewSubscriptionService(10)
	user1 := "did:example:user1"
	user2 := "did:example:user2"
	payload1 := NotificationPayload{ID: "4"}

	ch1, err := service.Subscribe(user1)
	require.NoError(t, err)
	ch2, err := service.Subscribe(user2)
	require.NoError(t, err)

	go service.Notify(user1, payload1)

	select {
	case received := <-ch1:
		require.Equal(t, payload1, received)
	case <-time.After(time.Second):
		t.Fatal("user1 should receive notification")
	}

	select {
	case <-ch2:
		t.Fatal("user2 should not receive notification")
	case <-time.After(100 * time.Millisecond):
		// Expected - user2 should not receive anything
	}
}
