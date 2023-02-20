package brutecli

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrEmptyCommand   = errors.New("command not specified")
	ErrUnkCommand     = errors.New("unknown command")
	ErrWrongArgsCount = errors.New("wrong arguments count")
)

type Command interface {
	GetName() string
	GetDesc() string
	Execute(context.Context, []string) (CmdResult, error)
}

type Commands map[string]Command

type CmdResult struct {
	Success bool
	Message string
	Code    int
}

func InitCommands(commands []Command) (Commands, error) {
	registry := make(Commands)
	for _, cmd := range commands {
		if err := registry.Add(cmd); err != nil {
			return nil, err
		}
	}
	return registry, nil
}

func (cs *Commands) Add(cmd Command) error {
	_, exists := (*cs)[cmd.GetName()]
	if exists {
		return fmt.Errorf("command %s already exists", cmd.GetName())
	}
	(*cs)[cmd.GetName()] = cmd
	return nil
}

func (cs Commands) Execute(ctx context.Context, args []string) (CmdResult, error) {
	if len(args) == 0 {
		return CmdResult{}, ErrEmptyCommand
	}
	cmdName := args[0]
	cmd, exists := cs[cmdName]
	if !exists {
		return CmdResult{}, fmt.Errorf("%s: %w", cmdName, ErrUnkCommand)
	}
	return cmd.Execute(ctx, args[1:])
}

func (cs Commands) Help() string {
	res := strings.Builder{}
	res.WriteString("Available Commands:\n")
	for _, command := range cs {
		res.WriteString(fmt.Sprintf(
			" - %s%s%s\n",
			command.GetName(), strings.Repeat(" ", 7-len(command.GetName())), command.GetDesc()),
		)
	}

	res.WriteString(" - quit   Exit\n")
	return res.String()
}

func ParseArgs(s string) []string {
	fields := strings.Fields(s)
	result := make([]string, 0, len(fields))
	for _, field := range fields {
		if field == "" {
			continue
		}
		result = append(result, field)
	}
	return result
}
