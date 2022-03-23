package main

import (
	"context"
	"github.com/mrinalirao/job-worker/cli"
	"log"
	"os"
)

func main() {

	params, err := cli.GetParams(os.Args[1:])
	if err != nil {
		log.Fatalf("Error parsing CLI parameters: %v", err)
	}
	//TODO: pass through a config file
	cfg := cli.NewClientConfig("cert/server-ca-cert.pem", "cert/userclient-key.pem", "cert/userclient-cert.pem")
	userClient, err := cli.NewWorkerClient(cfg)
	if err != nil {
		log.Fatalf("Error creating user client %v", err)
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
