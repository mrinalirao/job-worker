package cli

import "fmt"

type Params struct {
	CliCommand  string
	CommandName string
	Arguments   []string
	JobID       string
}

const (
	StartCmd  = "start"
	StopCmd   = "stop"
	StatusCmd = "status"
	StreamCmd = "stream"
)

func GetParams(args []string) (*Params, error) {
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

	if len(args) < 2 || args[0] != "-c" {
		return nil, fmt.Errorf("invalid parameters for %v command: %v", params.CliCommand, args)
	}

	params.CommandName = args[1]

	if len(args) >= 4 && args[2] == "-args" {
		params.Arguments = args[3:]
	}

	return &params, nil
}
