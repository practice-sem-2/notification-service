package server

import (
	"github.com/practice-sem-2/notification-service/internal/pb/notify"
	"github.com/practice-sem-2/notification-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationsServer struct {
	notify.UnimplementedNotificationsServer
	ucases *usecase.UseCase
}

func NewNotificationServer(ucases *usecase.UseCase) *NotificationsServer {
	return &NotificationsServer{
		ucases: ucases,
	}
}

func (s *NotificationsServer) Listen(r *notify.ListenRequest, server notify.Notifications_ListenServer) error {
	user, err := s.ucases.Verifier.GetUser(server.Context())

	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	listener := s.ucases.Notifications.Listen(user.Username)
	defer listener.Detach()
	for upd := range listener.Notifications() {
		notification := NotificationFromUpdate(upd)
		if notification != nil {
			err := server.Send(notification)
			if err != nil {
				return err
			}
		}

	}
	return nil
}
