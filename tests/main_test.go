package tests

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	common "github.com/vitermakov/otusgo-final/internal/app/config"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type MainSuiteTest struct {
	suite.Suite
	conn     *grpc.ClientConn
	client   pb.IPRuleClient
	eventIds []string
}

func (ms *MainSuiteTest) SetupTest() {
	cfg := common.Server{
		Host: "127.0.0.1",
		Port: 8088,
	}
	conn, err := grpc.Dial(
		net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	ms.Suite.Require().NoError(err)

	ms.client = pb.NewIPRuleClient(conn)
	ms.conn = conn
	ms.eventIds = make([]string, 0, 20)
}

func (ms *MainSuiteTest) TearDownTest() {
	_ = ms.conn.Close()
}

func (ms *MainSuiteTest) TestAddToWhiteList() {
	testCases := []struct {
		name         string
		white        bool
		ipNet        *pb.IPNet
		expectedCode codes.Code
	}{
		{
			name: "ok",
			ipNet: &pb.IPNet{
				IPNet: "192.168.1.0/24",
			},
			expectedCode: codes.OK,
		},
	}
	for _, tc := range testCases {
		ms.Suite.Run(tc.name, func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
			defer cancel()

			_, err := ms.client.AddToWhiteList(ctx, tc.ipNet)
			e, ok := status.FromError(err)
			ms.Suite.True(ok, "error is not status")
			ms.Suite.Equal(tc.expectedCode, e.Code())
		})
	}
}

func TestIPRule(t *testing.T) {
	suite.Run(t, new(MainSuiteTest))
}
