package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/Shopify/sarama"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/practice-sem-2/auth-tools"
	"github.com/practice-sem-2/notification-service/internal/pb/notify"
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
	"time"
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
	notify.RegisterNotificationsServer(grpcServer, server.NewNotificationServer(useCases, logger))

	return grpcServer, listener
}

func initUpdatesConsumers(logger *logrus.Logger) []storage.Consumer {
	brokers := strings.Split(viper.GetString("KAFKA_BROKERS"), ",")

	if len(brokers) == 0 {
		logger.Fatalf("KAFKA_BROKERS must be defined")
	}

	topicsList := viper.GetString("KAFKA_TOPICS")

	if topicsList == "" {
		logger.Fatalf("KAFKA_TOPICS must be defined")
	}

	topics := strings.Split(topicsList, ",")

	config := sarama.NewConfig()
	config.Consumer.Fetch.Max = 10
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	config.Consumer.Offsets.AutoCommit.Enable = false
	config.Consumer.Offsets.AutoCommit.Interval = 30 * time.Second
	consumers := make([]storage.Consumer, len(topics))
	for i, t := range topics {
		c, err := sarama.NewConsumer(brokers, config)
		if err != nil {
			logger.
				WithField("error", err.Error()).
				Fatalf("can't create consumer")
		}
		consumers[i] = storage.NewUpdatesConsumer(c, t, logger)
	}

	return consumers
}

func initNotificationStore(logger *logrus.Logger) *storage.NotificationStore {
	consumers := initUpdatesConsumers(logger)
	store := storage.NewNotificationStorage(logger, consumers...)
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

	publicKeyPath := viper.GetString("JWT_PUBLIC_KEY_PATH")
	verifier, err := auth.NewVerifierFromFile(publicKeyPath)
	if err != nil {
		logger.
			WithField("key_path", publicKeyPath).
			Fatalf("can't create verifier: %s", err.Error())
	}
	notificationUseCase := usecase.NewNotificationUseCase(store)
	useCases := usecase.NewUseCase(notificationUseCase, verifier)

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

	err = srv.Serve(lis)
	if err != nil {
		logger.Fatalf("grpc serving error: %s", err.Error())
	}
}
