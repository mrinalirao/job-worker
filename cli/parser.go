package cli

import (
	"flag"
	"fmt"
	"strings"
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
	fs := flag.NewFlagSet(command, flag.ExitOnError)
	jobID := fs.String("j", "", "job ID of the job")
	role := fs.String("r", "", "role to start the client with, options: [admin, user]")
	fs.Parse(args)
	if role == nil || (*role != "user" && *role != "admin") {
		return nil, fmt.Errorf("invalid role for %v command", params.CliCommand)
	}
	if jobID == nil {
		return nil, fmt.Errorf("missing jobID for %v command", params.CliCommand)
	}
	params.JobID = *jobID
	params.Role = *role

	return &params, nil
}

// Create a new type for a list of args
type argsList []string

// Implement the flag.Value interface
func (s *argsList) String() string {
	return fmt.Sprintf("%v", *s)
}

func (s *argsList) Set(value string) error {
	*s = strings.Split(value, ",")
	return nil
}

func getStartCommandParams(args []string) (*Params, error) {
	params := Params{
		CliCommand: StartCmd,
	}
	fs := flag.NewFlagSet(StartCmd, flag.ExitOnError)
	role := fs.String("r", "", "role to start the client with, options: [admin, user]")
	commandName := fs.String("c", "", "command to run")
	var a argsList
	fs.Var(&a, "args", "comma seperated list of args to the command")
	fs.Parse(args)
	if commandName == nil {
		return nil, fmt.Errorf("missing command name for %v command", params.CliCommand)
	}
	if role == nil || (*role != "user" && *role != "admin") {
		return nil, fmt.Errorf("invalid role for %v command", params.CliCommand)
	}
	params.Role = *role
	params.CommandName = *commandName
	params.Arguments = a

	return &params, nil
}
