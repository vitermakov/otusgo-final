package grpc

import (
	"context"
	"github.com/vitermakov/otusgo-final/pkg/utils/jsonx"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	common "github.com/vitermakov/otusgo-final/internal/app/config"
	config "github.com/vitermakov/otusgo-final/internal/app/config/brutefp"
	deps "github.com/vitermakov/otusgo-final/internal/app/deps/brutefp"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"github.com/vitermakov/otusgo-final/pkg/logger"
	"github.com/vitermakov/otusgo-final/pkg/utils/closer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IPRuleSuiteTest struct {
	suite.Suite
	closer *closer.Closer
	conn   *grpc.ClientConn
	client pb.IPRuleClient
	logger logger.Logger
}

func (is *IPRuleSuiteTest) SetupTest() {
	var err error
	cfg := getCfgAPI(is.T())

	logLevel, err := logger.ParseLevel(cfg.Logger.Level)
	is.Suite.Require().NoError(err)

	log, err := logger.NewLogrus(logger.Config{
		Level:    logLevel,
		FileName: cfg.Logger.FileName,
	})
	is.Suite.Require().NoError(err)
	is.logger = log

	is.closer = closer.NewCloser()

	// dbPool, closeFn := pgconn.NewPgConn(cfg.ServiceID, cfg.PgStore, log)
	// is.Suite.Require().NotNil(dbPool)
	// is.closer.Register("DB", closeFn)
	// sqlf.SetDialect(sqlf.PostgreSQL)

	repos, err := deps.NewRepos(cfg.Storage, nil)
	is.Suite.Require().NoError(err)

	depends := &deps.Deps{
		Repos:  repos,
		Logger: log,
	}

	services := deps.NewServices(depends, cfg)

	grpcServer, closeFn := NewHandledServer(cfg.API, services, depends)
	is.closer.Register("GRPC Server", closeFn)

	go func() {
		err := grpcServer.Start()
		is.Suite.Require().NoError(err)
	}()

	// сервер запускается не сразу
	for i := 0; i < 10; i++ {
		<-time.After(time.Millisecond * 10)
		is.conn, err = grpc.Dial(
			net.JoinHostPort(cfg.API.Host, strconv.Itoa(cfg.API.Port)),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			break
		}
	}
	is.Suite.Require().NoError(err)
	is.client = pb.NewIPRuleClient(is.conn)
}

func (is *IPRuleSuiteTest) TearDownTest() {
	_ = is.conn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	is.closer.Close(ctx, is.logger)
}

func (is *IPRuleSuiteTest) TestComplex() {
	type actionFn func(ctx context.Context, in *pb.IPNet, opts ...grpc.CallOption) (*emptypb.Empty, error)

	badNet := &pb.IPNet{IPNet: "292.168.1.0/24"}
	goodNet := &pb.IPNet{IPNet: "192.168.1.0/24"}

	testCases := []struct {
		name         string
		expectedCode codes.Code
		actionFn     actionFn
		arg          *pb.IPNet
	}{
		{
			name:         "white wrong net",
			expectedCode: codes.InvalidArgument,
			actionFn:     is.client.AddToWhiteList,
			arg:          badNet,
		}, {
			name:         "white ok",
			expectedCode: codes.OK,
			actionFn:     is.client.AddToWhiteList,
			arg:          goodNet,
		}, {
			name:         "white duplicate",
			expectedCode: codes.InvalidArgument,
			actionFn:     is.client.AddToWhiteList,
			arg:          goodNet,
		}, {
			name:         "white remove ok",
			expectedCode: codes.OK,
			actionFn:     is.client.DeleteFromWhiteList,
			arg:          goodNet,
		}, {
			name:         "white remove not found",
			expectedCode: codes.NotFound,
			actionFn:     is.client.DeleteFromWhiteList,
			arg:          goodNet,
		}, {
			name:         "black wrong net",
			expectedCode: codes.InvalidArgument,
			actionFn:     is.client.AddToBlackList,
			arg:          badNet,
		}, {
			name:         "black ok",
			expectedCode: codes.OK,
			actionFn:     is.client.AddToBlackList,
			arg:          goodNet,
		}, {
			name:         "black duplicate",
			expectedCode: codes.InvalidArgument,
			actionFn:     is.client.AddToBlackList,
			arg:          goodNet,
		}, {
			name:         "black remove ok",
			expectedCode: codes.OK,
			actionFn:     is.client.DeleteFromBlackList,
			arg:          goodNet,
		}, {
			name:         "black remove not found",
			expectedCode: codes.NotFound,
			actionFn:     is.client.DeleteFromBlackList,
			arg:          goodNet,
		},
	}

	for _, tc := range testCases {
		is.Suite.Run(tc.name, func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			// неверный адрес сети
			_, err := tc.actionFn(ctx, tc.arg)
			e, ok := status.FromError(err)

			is.Suite.True(ok, "error is not status")
			is.Suite.Equal(tc.expectedCode, e.Code())
		})
	}
}

func TestIPRuleApi(t *testing.T) {
	suite.Run(t, new(IPRuleSuiteTest))
}

func getCfgAPI(t *testing.T) config.Config {
	t.Helper()
	return config.Config{
		Logger: common.Logger{Level: "info"},
		API: common.Server{
			Host: "127.0.0.1",
			Port: 50051,
		},
		Limits: config.Limits{
			Method:         "fixed_memory",
			Store:          "memory",
			LoginPerMin:    10,
			PasswordPerMin: 20,
			IPPerMin:       30,
			BaseDuration:   jsonx.NewDuration(2, 's'),
		},
		Storage: common.Storage{
			Type: "memory",
			PGConn: common.SQLConn{
				Conn: common.Conn{
					Host:     "127.0.0.1",
					Port:     5432,
					User:     "otus_user",
					Password: "otus_pass",
				},
				DBName: "brutefp",
			},
		},
	}
}

func getConn(t *testing.T, cfgAPI common.Server) (*grpc.ClientConn, error) {
	t.Helper()
	var (
		err  error
		conn *grpc.ClientConn
	)
	// сервер запускается не сразу
	for i := 0; i < 10; i++ {
		<-time.After(time.Millisecond * 10)
		conn, err = grpc.Dial(
			net.JoinHostPort(cfgAPI.Host, strconv.Itoa(cfgAPI.Port)),
			grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			break
		}
	}
	return conn, err
}
