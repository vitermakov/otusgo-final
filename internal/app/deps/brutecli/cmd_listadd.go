package brutecli

import (
	"context"
	"fmt"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
)

type ListAdd struct {
	client pb.IPRuleClient
}

func (la ListAdd) GetName() string {
	return "add"
}

func (la ListAdd) GetDesc() string {
	return "Add network in list: add <white|black> <network>. Example: add white 192.168.2.0/24"
}

func (la ListAdd) Execute(ctx context.Context, args []string) error {
	if len(args) != 2 {
		return ErrWrongArgsCount
	}
	fmt.Printf("%s %s %s", la.GetName(), args[0], args[1])
	return nil
}

func NewListAdd(client pb.IPRuleClient) Command {
	return &ListAdd{client}
}
