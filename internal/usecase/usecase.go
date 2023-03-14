package usecase

type UseCase struct {
	Notifications *NotificationsUseCase
}

func NewUseCase(notifications *NotificationsUseCase) *UseCase {
	return &UseCase{
		Notifications: notifications,
	}
}
