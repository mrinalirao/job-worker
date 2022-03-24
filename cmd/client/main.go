package main

import (
	"context"
	"github.com/mrinalirao/job-worker/cli"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	params, err := cli.GetParams(os.Args[1:])
	if err != nil {
		logrus.Fatalf("Error parsing CLI parameters: %v", err)
	}
	var cfg cli.ClientConfig
	//TODO: pass through a config file
	if params.Role == "admin" {
		cfg = cli.NewClientConfig("cert/server-ca-cert.pem", "cert/adminclient-key.pem", "cert/adminclient-cert.pem")
	} else {
		cfg = cli.NewClientConfig("cert/server-ca-cert.pem", "cert/userclient-key.pem", "cert/userclient-cert.pem")
	}
	userClient, err := cli.NewWorkerClient(cfg)
	if err != nil {
		logrus.Fatalf("Error creating user client %v", err)
	}
	ctx := context.Background()
	switch params.CliCommand {
	case cli.StartCmd:
		if err := cli.StartJobHandler(ctx, userClient, params); err != nil {
			logrus.Fatalf("failed to run cmd: %v", err)
		}
	case cli.StopCmd:
		if err := cli.StopJobHandler(ctx, userClient, params); err != nil {
			logrus.Fatalf("failed to run cmd: %v", err)
		}
	case cli.StatusCmd:
		if err := cli.GetJobStatusHandler(ctx, userClient, params); err != nil {
			logrus.Fatalf("failed to run cmd: %v", err)
		}
	case cli.StreamCmd:
		if err := cli.GetJobOutputHandler(ctx, userClient, params); err != nil {
			logrus.Fatalf("failed to run cmd: %v", err)
		}
	}
}
