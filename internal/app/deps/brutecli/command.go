package brutecli

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrEmptyCommand   = errors.New("command not specified")
	ErrUnkCommand     = errors.New("unknown command")
	ErrWrongArgsCount = errors.New("wrong arguments count")
	ErrWrongArgument  = errors.New("wrong argument")
)

type Command interface {
	GetName() string
	GetDesc() string
	Execute(context.Context, []string) error
}

type Commands map[string]Command

func InitCommands([]Command) (Commands, error) {
	commands := make(Commands)
	for _, cmd := range commands {
		if err := commands.Add(cmd); err != nil {
			return nil, err
		}
	}
	return commands, nil
}

func (cs *Commands) Add(cmd Command) error {
	_, exists := (*cs)[cmd.GetName()]
	if exists {
		return fmt.Errorf("command %s already exists", cmd.GetName())
	}
	(*cs)[cmd.GetName()] = cmd
	return nil
}

func (cs Commands) Execute(ctx context.Context, args []string) error {
	if len(args) == 0 {
		return ErrEmptyCommand
	}
	cmdName := args[0]
	cmd, exists := cs[cmdName]
	if !exists {
		return fmt.Errorf("%s: %w", cmdName, ErrUnkCommand)
	}
	return cmd.Execute(ctx, args[1:])
}
