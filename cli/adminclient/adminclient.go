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
	//TODO: pass through a config file
	cfg := cli.NewClientConfig("cert/server-ca-cert.pem", "cert/adminclient-key.pem", "cert/adminclient-cert.pem")
	userClient, err := cli.NewWorkerClient(cfg)
	if err != nil {
		logrus.Fatalf("Error creating user client %v", err)
	}
	ctx := context.Background()
	switch params.CliCommand {
	case cli.StartCmd:
		cli.StartJobHandler(ctx, userClient, params)
	case cli.StopCmd:
		cli.StopJobHandler(ctx, userClient, params)
	case cli.StatusCmd:
		cli.GetJobStatusHandler(ctx, userClient, params)
	case cli.StreamCmd:
		cli.GetJobOutputHandler(ctx, userClient, params)
	}
}
