package usecase

import "github.com/practice-sem-2/auth-tools"

type UseCase struct {
	Verifier      *auth.VerifierService
	Notifications *NotificationsUseCase
}

func NewUseCase(notifications *NotificationsUseCase, verifier *auth.VerifierService) *UseCase {
	return &UseCase{
		Notifications: notifications,
		Verifier:      verifier,
	}
}
