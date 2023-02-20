package tests

import (
	"context"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
	config "github.com/vitermakov/otusgo-final/internal/app/config/brutefp"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type IPRuleSuiteTest struct {
	suite.Suite
	conn   *grpc.ClientConn
	client pb.IPRuleClient
}

func (is *IPRuleSuiteTest) SetupTest() {
	configFile := "../deployments/configs/brutefp_config.json"
	cfg, err := config.New(configFile)
	is.Suite.Require().NoError(err)

	conn, err := grpc.Dial(
		net.JoinHostPort(cfg.API.Host, strconv.Itoa(cfg.API.Port)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	is.Suite.Require().NoError(err)

	is.client = pb.NewIPRuleClient(conn)
	is.conn = conn
}

func (is *IPRuleSuiteTest) TearDownTest() {
	if is.conn != nil {
		err := is.conn.Close()
		is.Suite.Require().NoError(err)
	}
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

			is.Suite.Require().True(ok, "error is not status")
			is.Suite.Require().Equal(tc.expectedCode, e.Code())
		})
	}
}

func TestIPRule(t *testing.T) {
	suite.Run(t, new(IPRuleSuiteTest))
}
