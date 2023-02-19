package brutecli

import (
	"context"
	"fmt"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
)

type Reset struct {
	client pb.IPRuleClient
}

func (r Reset) GetName() string {
	return "reset"
}

func (r Reset) GetDesc() string {
	return "Reset bucket: reset <login|ip> <value>. Example: reset ip 192.168.0.1"
}

func (r Reset) Execute(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return ErrWrongArgsCount
	}
	fmt.Printf("%s %s %s", r.GetName(), args[0], args[1])
	return nil
}

func NewReset(client pb.IPRuleClient) Command {
	return &Reset{client}
}
