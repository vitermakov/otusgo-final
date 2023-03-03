package grpc

import (
	"context"

	"github.com/vitermakov/otusgo-final/internal/app/config"
	deps "github.com/vitermakov/otusgo-final/internal/app/deps/brutefp"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"github.com/vitermakov/otusgo-final/pkg/servers"
	grpcServ "github.com/vitermakov/otusgo-final/pkg/servers/grpc"
	"github.com/vitermakov/otusgo-final/pkg/utils/closer"
	"google.golang.org/grpc"
)

func NewHandledServer(
	config config.Server, services *deps.Services, deps *deps.Deps,
) (*grpcServ.Server, closer.CloseFunc) {
	server := grpcServ.NewServer(servers.NewConfig(
		config.Host,
		config.Port,
		false,
	), nil, deps.Logger)

	server.RegisterHandler(func(s *grpc.Server) {
		pb.RegisterIPRuleServer(s, IPRuleHandlerImpl{services: services, logger: deps.Logger})
		pb.RegisterPermitServer(s, PermitHandlerImpl{services: services, logger: deps.Logger})
	})

	return server, func(_ context.Context) error {
		server.Stop()
		return nil
	}
}
