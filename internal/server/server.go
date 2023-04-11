package server

import (
	"github.com/practice-sem-2/notification-service/internal/pb"
	"github.com/practice-sem-2/notification-service/internal/usecase"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationsServer struct {
	pb.UnimplementedNotificationsServer
	ucases *usecase.UseCase
}

func NewNotificationServer(ucases *usecase.UseCase) *NotificationsServer {
	return &NotificationsServer{
		ucases: ucases,
	}
}

func (s *NotificationsServer) Listen(r *pb.ListenRequest, server pb.Notifications_ListenServer) error {
	user, err := s.ucases.Verifier.GetUser(server.Context())

	if err != nil {
		return status.Error(codes.Unauthenticated, err.Error())
	}

	listener := s.ucases.Notifications.Listen(user.Username)
	for notification := range listener.Notifications() {
		err := server.Send(notification)
		if err != nil {
			listener.Detach()
			return err
		}
	}
	return nil
}
