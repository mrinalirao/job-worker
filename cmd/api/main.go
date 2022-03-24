package main

import (
	"github.com/mrinalirao/job-worker/server"
	"log"
)

func main() {
	if err := server.RunServer(); err != nil {
		log.Fatalf("failed to start server, %v", err)
	}
}
