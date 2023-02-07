package grpc

import (
	"errors"
	"net"
	"strconv"

	"github.com/vitermakov/otusgo-final/pkg/logger"
	"github.com/vitermakov/otusgo-final/pkg/servers"
	"google.golang.org/grpc"
)

type RegisterHandlerFunc func(s *grpc.Server)

type Server struct {
	*grpc.Server
	config      servers.Config
	Logger      logger.Logger
	AuthService servers.AuthService
}

func NewServer(config servers.Config, authSrv servers.AuthService, logger logger.Logger) *Server {
	unaryChain := grpc.ChainUnaryInterceptor(
		NewLoggerInterceptor(logger).Unary(),
		// NewAuthInterceptor(authSrv).Unary(),
	)
	return &Server{Server: grpc.NewServer(unaryChain), config: config, Logger: logger, AuthService: authSrv}
}

func (s *Server) RegisterHandler(handlerFunc RegisterHandlerFunc) {
	if handlerFunc == nil {
		return
	}
	handlerFunc(s.Server)
}

func (s *Server) Start() error {
	address := net.JoinHostPort(s.config.GetHost(), strconv.Itoa(s.config.GetPort()))
	socket, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	err = s.Server.Serve(socket)
	if err == nil || errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}
	return err
}

func (s *Server) Stop() {
	s.Server.GracefulStop()
}
