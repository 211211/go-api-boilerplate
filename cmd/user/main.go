package main

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/caarlos0/env"
	"github.com/vardius/go-api-boilerplate/pkg/common/application/jwt"
	"github.com/vardius/go-api-boilerplate/pkg/common/application/log"
	"github.com/vardius/go-api-boilerplate/pkg/common/application/os/shutdown"
	"github.com/vardius/go-api-boilerplate/pkg/common/infrastructure/commandbus/memory"
	"github.com/vardius/go-api-boilerplate/pkg/common/infrastructure/eventbus/memory"
	"github.com/vardius/go-api-boilerplate/pkg/common/infrastructure/eventstore/memory"
	"github.com/vardius/go-api-boilerplate/pkg/user/infrastructure/proto"
	server "github.com/vardius/go-api-boilerplate/pkg/user/interfaces/grpc"
	"google.golang.org/grpc"
)

type config struct {
	Env    string `env:"ENV"    envDefault:"development"`
	Host   string `env:"HOST"   envDefault:"localhost"`
	Port   int    `env:"PORT"   envDefault:"3001"`
	Secret string `env:"SECRET" envDefault:"secret"`
}

func main() {
	ctx := context.Background()

	cfg := config{}
	env.Parse(&cfg)

	logger := log.New(cfg.Env)
	jwtService := jwt.New([]byte(cfg.Secret), time.Hour*24)
	eventStore := eventstore.New()
	eventBus := eventbus.WithLogger("user", eventbus.New(), logger)
	commandBus := commandbus.WithLogger("user", commandbus.New(), logger)

	grpcServer := grpc.NewServer()
	userServer := server.New(
		commandBus,
		eventBus,
		eventStore,
		jwtService,
	)

	proto.RegisterUserServer(grpcServer, userServer)

	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	if err != nil {
		logger.Critical(ctx, "failed to listen: %v\n", err)
	} else {
		logger.Info(ctx, "[user] running at %s:%d\n", cfg.Host, cfg.Port)
	}

	go func() {
		logger.Critical(ctx, "failed to serve: %v\n", grpcServer.Serve(lis))
	}()

	shutdown.GracefulStop(func() {
		logger.Info(ctx, "[user] shutting down...\n")

		grpcServer.GracefulStop()

		logger.Info(ctx, "[user] gracefully stopped\n")
	})
}
