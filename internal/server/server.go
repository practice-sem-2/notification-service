package server

import (
	"github.com/practice-sem-2/notification-service/internal/pb/notify"
	"github.com/practice-sem-2/notification-service/internal/usecase"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationsServer struct {
	logger *logrus.Logger
	notify.UnimplementedNotificationsServer
	ucases *usecase.UseCase
}

func NewNotificationServer(ucases *usecase.UseCase, l *logrus.Logger) *NotificationsServer {
	return &NotificationsServer{
		ucases: ucases,
		logger: l,
	}
}

func (s *NotificationsServer) Listen(r *notify.ListenRequest, server notify.Notifications_ListenServer) error {
	s.logger.Infof("New listen request")
	user, err := s.ucases.Verifier.GetUser(server.Context())

	if err != nil {
		s.logger.Infof("User is authernticated. Aborting")
		return status.Error(codes.Unauthenticated, err.Error())
	}
	s.logger.Infof("Listening notifications for %s", user.Username)

	listener := s.ucases.Notifications.Listen(user.Username)

	for {
		select {
		case <-server.Context().Done():
			s.logger.Infof("User %s detached", user.Username)
			listener.Detach()
			return nil
		case upd := <-listener.Notifications():
			notification := NotificationFromUpdate(upd)
			if notification != nil {
				err := server.Send(notification)
				if err != nil {
					return err
				}
			}
		}
	}
}
