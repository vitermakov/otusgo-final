package brutecli

import (
	"context"
	"fmt"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"net"
)

// unexported
const (
	typeListWhite = "white"
	typeListBlack = "black"
)

type ListAdd struct {
	client pb.IPRuleClient
}

func (la ListAdd) GetName() string {
	return "add"
}

func (la ListAdd) GetDesc() string {
	return "Add network in list: add <type=white|black> <network>. Example: add white 192.168.2.0/24"
}

func (la ListAdd) Execute(ctx context.Context, args []string) (CmdResult, error) {
	return addOrRemoveExecute(ctx, la.client, args, true)
}

func NewListAdd(client pb.IPRuleClient) Command {
	return &ListAdd{client}
}

func addOrRemoveExecute(ctx context.Context, client pb.IPRuleClient, args []string, bAdd bool) (CmdResult, error) {
	if len(args) != 2 {
		return CmdResult{}, ErrWrongArgsCount
	}
	tip := args[0]
	nw := args[1]
	if tip != typeListWhite && tip != typeListBlack {
		return CmdResult{}, fmt.Errorf("type must be %s or %s", typeListWhite, typeListBlack)
	}
	_, _, err := net.ParseCIDR(args[1])
	if err != nil {
		return CmdResult{}, err
	}
	network := &pb.IPNet{IPNet: nw}
	if bAdd {
		if tip == typeListWhite {
			_, err = client.AddToWhiteList(ctx, network)
		} else {
			_, err = client.AddToBlackList(ctx, network)
		}
	} else {
		if tip == typeListWhite {
			_, err = client.DeleteFromWhiteList(ctx, network)
		} else {
			_, err = client.DeleteFromBlackList(ctx, network)
		}
	}
	return makeResult(err)
}

func makeResult(err error) (CmdResult, error) {
	s, ok := status.FromError(err)
	res := CmdResult{Message: s.Message(), Code: int(s.Code())}
	if ok {
		if s.Code() == codes.OK {
			res.Success = true
		}
	}
	return res, nil
}
