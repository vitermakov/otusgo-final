package app

import (
	"bufio"
	"context"
	"fmt"
	"github.com/vitermakov/otusgo-final/internal/app/deps/brutecli"
	"net"
	"os"
	"strconv"

	config "github.com/vitermakov/otusgo-final/internal/app/config/brutefp-cli"
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

func NewBruteFPCli(config config.Config) App {
	return &BruteFPCli{config: config}
}

func (cli *BruteFPCli) Initialize(ctx context.Context) error {
	cliCfg := cli.config.GrpcClient
	conn, err := grpc.DialContext(ctx,
		net.JoinHostPort(cliCfg.Host, strconv.Itoa(cliCfg.Port)),
		grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("can't connect to grpc server on %s:%d: %w", cliCfg.Host, cliCfg.Port, err)
	}

	cli.irClient = pb.NewIPRuleClient(conn)
	cli.pmClient = pb.NewPermitClient(conn)
	cli.conn = conn

	cli.commands, err = brutecli.InitCommands([]brutecli.Command{
		brutecli.NewListAdd(cli.irClient),
		brutecli.NewListRm(cli.irClient),
		brutecli.NewReset(cli.pmClient),
	})
	if err != nil {
		return fmt.Errorf("can't init commands: %w", err)
	}

	return nil
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
				fmt.Printf("error: %s", err)
				continue
			}
			if result.Success {
				fmt.Printf("OK: %s", result.Message)
			} else {
				fmt.Printf("Error: %s", result.Message)
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