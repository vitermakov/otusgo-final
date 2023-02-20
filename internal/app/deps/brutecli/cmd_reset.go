package brutecli

import (
	"context"
	"fmt"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"github.com/vitermakov/otusgo-final/internal/model"
)

type Reset struct {
	client pb.PermitClient
}

func (r Reset) GetName() string {
	return "reset"
}

func (r Reset) GetDesc() string {
	return "Reset bucket: reset <login|ip> <value>. Example: reset ip 192.168.0.1"
}

func (r Reset) Execute(ctx context.Context, args []string) (CmdResult, error) {
	var err error
	if len(args) != 2 {
		return CmdResult{}, ErrWrongArgsCount
	}
	param := args[0]
	value := args[1]
	switch param {
	case model.LimitParamNameLogin:
		_, err = r.client.ResetLogin(ctx, &pb.RstLoginReq{Login: value})
	case model.LimitParamNameIP:
		_, err = r.client.ResetIP(ctx, &pb.RstIPReq{IP: value})
	default:
		return CmdResult{}, fmt.Errorf(
			"reset param must be %s or %s", model.LimitParamNameLogin, model.LimitParamNameIP,
		)
	}
	return makeResult(err)
}

func NewReset(client pb.PermitClient) Command {
	return &Reset{client}
}
