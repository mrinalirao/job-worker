package cli

import (
	"fmt"
)

type Params struct {
	CliCommand  string
	CommandName string
	Arguments   []string
	JobID       string
	Role        string
}

const (
	StartCmd  = "start"
	StopCmd   = "stop"
	StatusCmd = "status"
	StreamCmd = "stream"
)

func GetParams(args []string) (*Params, error) {
	//TODO: parse env vars using flag
	argsLen := len(args)

	if argsLen < 2 {
		return nil, fmt.Errorf("invalid parameters %v", args)
	}

	switch args[0] {
	case StartCmd:
		return getStartCommandParams(args[1:])
	case StopCmd:
		return getJobCommandParams(StopCmd, args[1:])
	case StatusCmd:
		return getJobCommandParams(StatusCmd, args[1:])
	case StreamCmd:
		return getJobCommandParams(StreamCmd, args[1:])
	}

	return nil, fmt.Errorf("invalid command %v", args)
}

func getJobCommandParams(command string, args []string) (*Params, error) {
	params := Params{
		CliCommand: command,
	}

	if len(args) != 2 || args[0] != "-j" {
		return nil, fmt.Errorf("invalid parameters for %v command: %v", params.CliCommand, args)
	}
	params.JobID = args[1]

	return &params, nil
}

func getStartCommandParams(args []string) (*Params, error) {
	params := Params{
		CliCommand: StartCmd,
	}

	if len(args) < 4 || args[0] != "-r" || args[2] != "-c" {
		return nil, fmt.Errorf("invalid parameters for %v command: %v", params.CliCommand, args)
	}

	params.Role = args[1]
	params.CommandName = args[3]

	if len(args) >= 6 && args[4] == "-args" {
		params.Arguments = args[5:]
	}

	return &params, nil
}
