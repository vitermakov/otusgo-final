package app

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"os"
	"strconv"

	config "github.com/vitermakov/otusgo-final/internal/app/config/brutefp-cli"
	"github.com/vitermakov/otusgo-final/internal/app/deps/brutecli"
	"github.com/vitermakov/otusgo-final/internal/handler/grpc/pb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BruteFPCli struct {
	config   config.Config
	commands brutecli.Commands
	conn     *grpc.ClientConn
	irClient pb.IPRuleClient
	pmClient pb.PermitClient
}

func NewBruteFPCli(ctx context.Context, config config.Config) (App, error) {
	cliCfg := config.GrpcClient
	conn, err := grpc.DialContext(ctx,
		net.JoinHostPort(cliCfg.Host, strconv.Itoa(cliCfg.Port)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("can't connect to grpc server on %s:%d: %w", cliCfg.Host, cliCfg.Port, err)
	}

	irClient := pb.NewIPRuleClient(conn)
	pmClient := pb.NewPermitClient(conn)

	commands, err := brutecli.InitCommands([]brutecli.Command{
		brutecli.NewListAdd(irClient),
		brutecli.NewListRm(irClient),
		brutecli.NewReset(pmClient),
	})
	if err != nil {
		return nil, fmt.Errorf("can't init commands: %w", err)
	}
	return &BruteFPCli{
		config:   config,
		conn:     conn,
		irClient: irClient,
		pmClient: pmClient,
		commands: commands,
	}, nil
}

func (cli *BruteFPCli) Run(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cli.printHelp()

	// Входим в интерактивный режим приема команд.
	go func() {
		defer func() {
			cancel()
		}()

		reader := bufio.NewScanner(os.Stdin)
		for {
			fmt.Print("> ")
			if !reader.Scan() {
				break
			}
			cmdLine := reader.Text()
			args := brutecli.ParseArgs(cmdLine)
			if cmdLine == "quit" {
				return
			}
			result, err := cli.commands.Execute(ctx, args)
			if err != nil {
				fmt.Printf("error: %s\n", err)
				continue
			}
			if result.Success {
				fmt.Print("OK\n")
			} else {
				fmt.Printf("Error: %s\n", result.Message)
			}
		}
	}()

	<-ctx.Done()

	return nil
}

func (cli *BruteFPCli) Close() {
	if err := cli.conn.Close(); err != nil {
		fmt.Printf("error closing grpc connection: %s\n", err.Error())
	}
	fmt.Println("\nBruteFP CLI quit. Bye!")
}

func (cli *BruteFPCli) printHelp() {
	fmt.Printf("%s\n\n%s\n", cli.config.ServiceName, cli.commands.Help())
}
