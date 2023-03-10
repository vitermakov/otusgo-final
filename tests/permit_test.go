package tests

import (
	"context"
	"fmt"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	config "github.com/vitermakov/otusgo-final/internal/app/config/brutefp"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"github.com/vitermakov/otusgo-final/internal/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type PermitSuiteTest struct {
	suite.Suite
	config   config.Config
	conn     *grpc.ClientConn
	irClient pb.IPRuleClient
	pmClient pb.PermitClient
}

func (ps *PermitSuiteTest) SetupTest() {
	configFile := "/app/deployments/configs/brutefp_config.json"
	var err error

	ps.config, err = config.New(configFile)
	ps.Suite.Require().NoError(err)

	conn, err := grpc.Dial(
		net.JoinHostPort(ps.config.API.Host, strconv.Itoa(ps.config.API.Port)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	ps.Suite.Require().NoError(err)

	ps.irClient = pb.NewIPRuleClient(conn)
	ps.pmClient = pb.NewPermitClient(conn)
	ps.conn = conn
}

func (ps *PermitSuiteTest) TearDownTest() {
	if ps.conn != nil {
		err := ps.conn.Close()
		ps.Suite.Require().NoError(err)
	}
}

// TestOutOfLimits тестируем не только ограничение по количеству запросов по каждому
// отдельному login/password/ip, но и то, что разные типы ограничений работают раздельно
// друг от друга. Дальнейшие тесты будут выполняться на проверке одного параметра (например, login).
func (ps *PermitSuiteTest) TestOutOfLimits() {
	cfg := ps.config
	limits := []struct {
		param string
		limit int
	}{
		{
			param: model.LimitParamNameLogin,
			limit: cfg.Limits.LoginPerMin,
		}, {
			param: model.LimitParamNamePassword,
			limit: cfg.Limits.PasswordPerMin,
		}, {
			param: model.LimitParamNameIP,
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
				Login:    fmt.Sprintf("login_lims_%d", i),
				Password: fmt.Sprintf("password_lims_%d", i),
				IP:       fmt.Sprintf("192.168.0.%d", i),
			}
			switch limit.param {
			case model.LimitParamNameLogin:
				req.Login = "login_lims"
			case model.LimitParamNamePassword:
				req.Password = "password_lims"
			case model.LimitParamNameIP:
				req.IP = "192.168.1.1"
			}
			res, err := ps.pmClient.CheckQuery(ctx, req)
			ps.Suite.Require().NoError(err)

			if i <= limit.limit {
				ps.Suite.Require().True(res.Success, "%s limit is %d: got %d", limit.param, limit.limit, i)
			} else {
				// проверяем не только флаг Success, но им причину
				ps.Suite.Require().False(res.Success, "%s limit is %d: got %d", limit.param, limit.limit, i)
				switch limit.param {
				case model.LimitParamNameLogin:
					ps.Suite.Require().Contains(res.GetReason(), model.ErrDeniedByLoginLimit.Error())
				case model.LimitParamNamePassword:
					ps.Suite.Require().Contains(res.GetReason(), model.ErrDeniedByPasswordLimit.Error())
				case model.LimitParamNameIP:
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
		Login:    "login_wl",
		Password: "password_wl",
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
	cfg := ps.config
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	_, err := ps.irClient.AddToWhiteList(ctx, &pb.IPNet{IPNet: "192.168.4.0/24"})
	ps.Suite.Require().NoError(err)

	for i := 1; i <= 2*cfg.Limits.LoginPerMin; i++ {
		req := &pb.PermitReq{
			Login:    "login_bl",
			Password: fmt.Sprintf("password_bl_%d", i),
			IP:       fmt.Sprintf("192.168.4.%d", i),
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
	cfg := ps.config
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// добираемся до лимита
	for i := 1; i <= cfg.Limits.LoginPerMin; i++ {
		req := &pb.PermitReq{
			Login:    "login_rs",
			Password: fmt.Sprintf("password_rs_%d", i),
			IP:       fmt.Sprintf("192.168.5.%d", i),
		}
		res, err := ps.pmClient.CheckQuery(ctx, req)
		ps.Suite.Require().NoError(err)
		ps.Suite.Require().True(res.Success)
	}

	// сбрасываем бакет с логином
	_, err := ps.pmClient.ResetLogin(ctx, &pb.RstLoginReq{Login: "login_rs"})
	ps.Suite.Require().NoError(err)

	// убеждаемся что следующие запросы разрешены
	for i := 1; i <= cfg.Limits.LoginPerMin; i++ {
		req := &pb.PermitReq{
			Login:    "login_rs",
			Password: fmt.Sprintf("password_rs_%d", i),
			IP:       fmt.Sprintf("192.168.5.%d", i),
		}
		res, err := ps.pmClient.CheckQuery(ctx, req)
		ps.Suite.Require().NoError(err)
		ps.Suite.Require().True(res.Success)
	}
}

// TestAutoReset убеждаемся, что ограничение работает не вечно, а только в рамках периода.
func (ps *PermitSuiteTest) TestAutoReset() {
	cfg := ps.config
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// добираемся до лимита + 1
	for i := 1; i <= cfg.Limits.LoginPerMin+1; i++ {
		req := &pb.PermitReq{
			Login:    "login_ar",
			Password: fmt.Sprintf("password_ar_%d", i),
			IP:       fmt.Sprintf("192.168.6.%d", i),
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
		Login:    "login_ar",
		Password: "password_ar_1",
		IP:       "192.168.6.1",
	}
	res, err := ps.pmClient.CheckQuery(ctx, req)
	ps.Suite.Require().NoError(err)
	ps.Suite.Require().True(res.Success)
}

func TestPermitApi(t *testing.T) {
	suite.Run(t, new(PermitSuiteTest))
}
