package server

import (
	"github.com/practice-sem-2/notification-service/internal/pb"
	"github.com/practice-sem-2/notification-service/internal/usecase"
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
	panic("Unimplemented")
}
