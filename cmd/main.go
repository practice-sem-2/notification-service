package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/practice-sem-2/notification-service/internal/pb"
	"github.com/practice-sem-2/notification-service/internal/server"
	"github.com/practice-sem-2/notification-service/internal/storage"
	"github.com/practice-sem-2/notification-service/internal/usecase"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
)

func initLogger(level string) *logrus.Logger {

	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{
		PrettyPrint: true,
	})

	logLevel, err := logrus.ParseLevel(level)
	if err != nil {
		logger.SetLevel(logrus.InfoLevel)
		logger.
			WithField("log_level", level).
			Warning("specified invalid log level")
	} else {
		logger.SetLevel(logLevel)
		logger.
			WithField("log_level", level).
			Infof("specified %s log level", logLevel.String())
	}

	return logger
}

func initServer(address string, useCases *usecase.UseCase, logger *logrus.Logger) (*grpc.Server, net.Listener) {

	listener, err := net.Listen("tcp", address)
	logger.Infof("start listening on %s", address)

	if err != nil {
		logger.Fatalf("can't listen to address: %s", err.Error())
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNotificationsServer(grpcServer, server.NewNotificationServer(useCases))

	return grpcServer, listener
}

func initConsumerGroup(logger *logrus.Logger) sarama.ConsumerGroup {
	brokers := strings.Split(viper.GetString("KAFKA_BROKERS"), ",")

	if len(brokers) == 0 {
		logger.Fatalf("KAFKA_BROKERS must be defined")
	}

	topic := viper.GetString("KAFKA_NOTIFICATIONS_TOPIC")

	if topic == "" {
		logger.Fatalf("KAFKA_NOTIFICATIONS_TOPIC must be defined")
	}

	config := sarama.NewConfig()
	group, err := sarama.NewConsumerGroup(brokers, topic, config)

	if err != nil {
		logger.
			WithField("error", err.Error()).
			Fatalf("can't create consumer group")
	}

	return group
}

func initNotificationStore(logger *logrus.Logger) *storage.NotificationStore {
	group := initConsumerGroup(logger)
	store := storage.NewNotificationStorage(group, logger)
	return store
}

func main() {
	viper.AutomaticEnv()
	ctx := context.Background()
	defer ctx.Done()

	var host string
	var port int
	var logLevel string

	flag.IntVar(&port, "port", 80, "port on which server will be started")
	flag.StringVar(&host, "host", "0.0.0.0", "host on which server will be started")
	flag.StringVar(&logLevel, "log", "info", "log level")

	flag.Parse()

	logger := initLogger(logLevel)
	store := initNotificationStore(logger)
	go func() {
		err := store.Run(ctx)
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.
				WithField("error", err).
				Error("store listening ended with error")
		}
	}()

	notificationUseCase := usecase.NewNotificationUseCase(store)
	useCases := usecase.NewUseCase(notificationUseCase)

	address := fmt.Sprintf("%s:%d", host, port)
	srv, lis := initServer(address, useCases, logger)
	osSignal := make(chan os.Signal, 1)
	signal.Notify(osSignal,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)

	go func(ctx context.Context) {
		select {
		case sig := <-osSignal:
			srv.GracefulStop()
			logger.Infof("%s caught. Gracefully shutdown", sig.String())
		case <-ctx.Done():
			return
		}
	}(ctx)

	err := srv.Serve(lis)
	if err != nil {
		logger.Fatalf("grpc serving error: %s", err.Error())
	}
}
