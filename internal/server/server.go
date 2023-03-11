package server

import (
	"github.com/practice-sem-2/notification-service/internal/pb"
	"github.com/practice-sem-2/notification-service/internal/usecase"
)

type NotificationsServer struct {
	pb.UnimplementedNotificationsServer
}

func NewNotificationServer(ucase *usecase.UseCase) *NotificationsServer {
	//TODO implement me
	panic("implement me")
}

func (n *NotificationsServer) Listen(r *pb.ListenRequest, server pb.Notifications_ListenServer) error {
	//TODO implement me
	panic("implement me")
}
