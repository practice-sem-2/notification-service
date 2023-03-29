package usecase

import "github.com/practice-sem-2/auth-tools"

type UseCase struct {
	Verifier *auth.VerifierService
}

func NewUseCase(verifier *auth.VerifierService) *UseCase {
	return &UseCase{
		Verifier: verifier,
	}
}
