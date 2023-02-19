package brutecli

import (
	"context"
	"fmt"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
)

type ListRm struct {
	client pb.IPRuleClient
}

func (lr ListRm) GetName() string {
	return "rm"
}

func (lr ListRm) GetDesc() string {
	return "Remove network from list: rm <white|black> <network>. Example: rm black 192.168.2.0/24"
}

func (lr ListRm) Execute(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return ErrWrongArgsCount
	}
	fmt.Printf("%s %s %s", lr.GetName(), args[0], args[1])
	return nil
}

func NewListRm(client pb.IPRuleClient) Command {
	return &ListRm{client}
}
