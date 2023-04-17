package storage

import (
	"context"
	"encoding/json"
	"github.com/practice-sem-2/notification-service/internal/models"
	"github.com/sirupsen/logrus"
	"github.com/zyedidia/generic/multimap"
	"strings"
	"sync"
)

// TODO IDK how to correctly choose channel buffer size
const readerBufferSize = 16

type Worker func()

type NotificationListener struct {
	UserID   string
	store    *NotificationStore
	listener chan models.Update
}

func (l *NotificationListener) Notifications() <-chan models.Update {
	return l.listener
}

// Detach cancels listening and closes the listener channel
func (l *NotificationListener) Detach() {
	l.store.detach(l)
}

type Consumer interface {
	Run(ctx context.Context, updates chan<- models.Update) error
}

type NotificationStore struct {
	rm        sync.RWMutex
	consumers []Consumer
	listeners multimap.MultiMap[string, chan models.Update]
	logger    *logrus.Logger
}

func NewNotificationStorage(logger *logrus.Logger, consumers ...Consumer) *NotificationStore {
	store := &NotificationStore{
		consumers: consumers,
		listeners: multimap.NewMapSlice[string, chan models.Update](),
		logger:    logger,
	}
	return store
}

func (s *NotificationStore) Notify(userID string, msg models.Update) {
	data, _ := json.Marshal(msg)
	s.logger.
		WithField("update", string(data)).
		Infof("Notifying %s", userID)
	s.rm.RLock()
	for _, reader := range s.listeners.Get(userID) {
		reader <- msg
	}
	defer s.rm.RUnlock()
}

func (s *NotificationStore) detach(listener *NotificationListener) {
	s.rm.Lock()
	defer s.rm.Unlock()
	s.listeners.Remove(listener.UserID, listener.listener)
	close(listener.listener)
	s.logger.Infof("Listener of %s detached", listener.UserID)
}

func (s *NotificationStore) Run(ctx context.Context) error {
	s.logger.Info("Running the store")
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	upds := make(chan models.Update, readerBufferSize)
	go s.fanOutUpdates(ctx, upds)
	for _, cons := range s.consumers {
		wg.Add(1)
		go func(c Consumer) {
			err := c.Run(ctx, upds)

			if err != nil {
				s.logger.Errorf("one of consumers failed with error: %v", err)
			}
		}(cons)
	}

	wg.Wait()
	close(upds)
	return nil
}

func (s *NotificationStore) fanOutUpdates(ctx context.Context, upds chan models.Update) {
	for {
		select {
		case _ = <-ctx.Done():
			break
		case upd, ok := <-upds:
			if !ok {
				s.logger.Warn("Updates channel closed. Stop notifying clients")
				break
			}
			s.logger.Infof("New updates for audience: %s", strings.Join(upd.GetAudience(), ","))
			for _, dest := range upd.GetAudience() {
				s.logger.Infof("Notifying %s", dest)
				s.Notify(dest, upd)
			}
		}
	}
}

// Listen returns a channel contains all notification connected to userID.
func (s *NotificationStore) Listen(userID string) NotificationListener {
	s.rm.Lock()
	defer s.rm.Unlock()
	listener := make(chan models.Update, readerBufferSize)
	s.listeners.Put(userID, listener)
	s.logger.Infof("Created listener for %s", userID)
	return NotificationListener{
		UserID:   userID,
		store:    s,
		listener: listener,
	}
}
