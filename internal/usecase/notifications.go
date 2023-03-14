package usecase

import "github.com/practice-sem-2/notification-service/internal/storage"

type NotificationsUseCase struct {
	store *storage.NotificationStore
}

func NewNotificationUseCase(store *storage.NotificationStore) *NotificationsUseCase {
	return &NotificationsUseCase{
		store: store,
	}
}

func (u *NotificationsUseCase) Listen(userID string) storage.NotificationListener {
	return u.store.Listen(userID)
}
