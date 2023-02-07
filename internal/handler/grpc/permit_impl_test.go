package grpc

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	deps "github.com/vitermakov/otusgo-final/internal/app/deps/brutefp"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"github.com/vitermakov/otusgo-final/internal/model"
	"github.com/vitermakov/otusgo-final/internal/ratelimit"
	"github.com/vitermakov/otusgo-final/pkg/logger"
	"github.com/vitermakov/otusgo-final/pkg/utils/closer"
	"google.golang.org/grpc"
)

type PermitSuiteTest struct {
	suite.Suite
	closer   *closer.Closer
	irConn   *grpc.ClientConn
	pmConn   *grpc.ClientConn
	irClient pb.IPRuleClient
	pmClient pb.PermitClient
	logger   logger.Logger
	services *deps.Services
}

func (ps *PermitSuiteTest) SetupTest() {
	var err error
	cfg := getCfgAPI(ps.T())

	logLevel, err := logger.ParseLevel(cfg.Logger.Level)
	ps.Suite.Require().NoError(err)

	log, err := logger.NewLogrus(logger.Config{
		Level:    logLevel,
		FileName: cfg.Logger.FileName,
	})
	ps.Suite.Require().NoError(err)
	ps.logger = log

	ps.closer = closer.NewCloser()

	// dbPool, closeFn := pgconn.NewPgConn(cfg.ServiceID, cfg.PgStore, log)
	// ps.Suite.Require().NotNil(dbPool)
	// ps.closer.Register("DB", closeFn)
	// sqlf.SetDialect(sqlf.PostgreSQL)

	rateLimiter, closeFn, err := ratelimit.NewRateLimiter(cfg.Limits)
	ps.Suite.Require().NoError(err)
	ps.closer.Register("Rate Limiter", closeFn)

	repos, err := deps.NewRepos(cfg.Storage, nil)
	ps.Suite.Require().NoError(err)

	depends := &deps.Deps{
		Repos:       repos,
		Logger:      log,
		RateLimiter: rateLimiter,
	}

	ps.services = deps.NewServices(depends, cfg)

	grpcServer, closeFn := NewHandledServer(cfg.API, ps.services, depends)
	ps.closer.Register("GRPC Server", closeFn)

	go func() {
		err := grpcServer.Start()
		ps.Suite.Require().NoError(err)
	}()

	ps.irConn, err = getConn(ps.T(), cfg.API)
	ps.Suite.Require().NoError(err)
	ps.pmConn, err = getConn(ps.T(), cfg.API)
	ps.Suite.Require().NoError(err)
	ps.irClient = pb.NewIPRuleClient(ps.irConn)
	ps.pmClient = pb.NewPermitClient(ps.pmConn)
}

func (ps *PermitSuiteTest) TearDownTest() {
	_ = ps.irConn.Close()
	_ = ps.pmConn.Close()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	ps.closer.Close(ctx, ps.logger)
}

// TestOutOfLimits тестируем не только ограничение по количеству запросов по каждому
// отдельному login/password/ip, но и то, что разные типы ограничений работают раздельно
// друг от друга. Дальнейшие тесты будут выполняться на проверке одного параметра (например, login).
func (ps *PermitSuiteTest) TestOutOfLimits() {
	cfg := getCfgAPI(ps.T())
	limits := []struct {
		param string
		limit int
	}{
		{
			param: "login",
			limit: cfg.Limits.LoginPerMin,
		}, {
			param: "password",
			limit: cfg.Limits.PasswordPerMin,
		}, {
			param: "ip",
			limit: cfg.Limits.IPPerMin,
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	for _, limit := range limits {
		// половина - permit, остальное - not permit
		for i := 1; i <= limit.limit*2; i++ {
			// проверяемому параметру даем одно значение, остальным разные
			req := &pb.PermitReq{
				Login:    fmt.Sprintf("login_%d", i),
				Password: fmt.Sprintf("password_%d", i),
				IP:       fmt.Sprintf("192.168.0.%d", i),
			}
			switch limit.param {
			case "login":
				req.Login = limit.param
			case "password":
				req.Password = limit.param
			case "ip":
				req.IP = "192.168.1.1"
			}
			res, err := ps.pmClient.CheckQuery(ctx, req)
			ps.Suite.Require().NoError(err)

			if i <= limit.limit {
				ps.Suite.Require().True(res.Success, "%s limit is %d: got %d", limit.param, limit.limit, i)
			} else {
				// проверяем не только флаг Success, но им причину
				ps.Suite.Require().False(res.Success)
				switch limit.param {
				case "login":
					ps.Suite.Require().Contains(res.GetReason(), model.ErrDeniedByLoginLimit.Error())
				case "password":
					ps.Suite.Require().Contains(res.GetReason(), model.ErrDeniedByPasswordLimit.Error())
				case "ip":
					ps.Suite.Require().Contains(res.GetReason(), model.ErrDeniedByIPLimit.Error())
				}
			}
		}
	}
}

// TestBlackList если добавить IP в black-list, то не разрешен ни один запрос.
func (ps *PermitSuiteTest) TestBlackList() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := ps.irClient.AddToBlackList(ctx, &pb.IPNet{IPNet: "192.168.3.0/24"})
	ps.Suite.Require().NoError(err)

	req := &pb.PermitReq{
		Login:    "login",
		Password: "password",
		IP:       "192.168.3.100",
	}
	res, err := ps.pmClient.CheckQuery(ctx, req)
	ps.Suite.Require().NoError(err)

	ps.Suite.Require().False(res.Success)
	ps.Suite.Require().Contains(res.GetReason(), model.ErrDeniedByRule.Error())

	_, err = ps.irClient.DeleteFromBlackList(ctx, &pb.IPNet{IPNet: "192.168.3.0/24"})
	ps.Suite.Require().NoError(err)
}

// TestWhiteList если добавить IP в white-list, то ограничение по кол-ву запросов не действует.
func (ps *PermitSuiteTest) TestWhiteList() {
	cfg := getCfgAPI(ps.T())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := ps.irClient.AddToWhiteList(ctx, &pb.IPNet{IPNet: "192.168.4.0/24"})
	ps.Suite.Require().NoError(err)

	for i := 1; i <= 2*cfg.Limits.LoginPerMin; i++ {
		req := &pb.PermitReq{
			Login:    "login",
			Password: fmt.Sprintf("password_%d", i),
			IP:       "192.168.4.100",
		}
		res, err := ps.pmClient.CheckQuery(ctx, req)
		ps.Suite.Require().NoError(err)
		ps.Suite.Require().True(res.Success)
	}

	_, err = ps.irClient.DeleteFromWhiteList(ctx, &pb.IPNet{IPNet: "192.168.4.0/24"})
	ps.Suite.Require().NoError(err)
}

// TestBucketReset проверяем сброс бакета (по логину).
func (ps *PermitSuiteTest) TestBucketReset() {
	cfg := getCfgAPI(ps.T())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// добираемся до лимита
	for i := 1; i <= cfg.Limits.LoginPerMin; i++ {
		req := &pb.PermitReq{
			Login:    "login",
			Password: fmt.Sprintf("password_%d", i),
			IP:       fmt.Sprintf("192.168.5.%d", i),
		}
		res, err := ps.pmClient.CheckQuery(ctx, req)
		ps.Suite.Require().NoError(err)
		ps.Suite.Require().True(res.Success)
	}

	// сбрасываем бакет с логином
	_, err := ps.pmClient.ResetLogin(ctx, &pb.RstLoginReq{Login: "login"})
	ps.Suite.Require().NoError(err)

	// убеждаемся что следующие запросы разрешены
	for i := 1; i <= cfg.Limits.LoginPerMin; i++ {
		req := &pb.PermitReq{
			Login:    "login",
			Password: fmt.Sprintf("password_%d", i),
			IP:       fmt.Sprintf("192.168.5.%d", i),
		}
		res, err := ps.pmClient.CheckQuery(ctx, req)
		ps.Suite.Require().NoError(err)
		ps.Suite.Require().True(res.Success)
	}
}

// TestAutoReset убеждаемся, что ограничение работает не вечно, а только в рамках периода.
func (ps *PermitSuiteTest) TestAutoReset() {
	cfg := getCfgAPI(ps.T())
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	ps.services.PermitChecker.SetBaseDuration(time.Second * 2)

	// добираемся до лимита + 1
	for i := 1; i <= cfg.Limits.LoginPerMin+1; i++ {
		req := &pb.PermitReq{
			Login:    "login",
			Password: fmt.Sprintf("password_%d", i),
			IP:       fmt.Sprintf("192.168.5.%d", i),
		}
		res, err := ps.pmClient.CheckQuery(ctx, req)
		ps.Suite.Require().NoError(err)
		if i == cfg.Limits.LoginPerMin+1 {
			ps.Suite.Require().False(res.Success)
		} else {
			ps.Suite.Require().True(res.Success)
		}
	}
	// ждем период
	<-time.After(time.Second * 2)

	// запросы опять проходят
	req := &pb.PermitReq{
		Login:    "login",
		Password: "password_1",
		IP:       "192.168.5.1",
	}
	res, err := ps.pmClient.CheckQuery(ctx, req)
	ps.Suite.Require().NoError(err)
	ps.Suite.Require().True(res.Success)
}

func TestPermitApi(t *testing.T) {
	suite.Run(t, new(PermitSuiteTest))
}
