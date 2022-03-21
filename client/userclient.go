package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/mrinalirao/job-worker/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"io/ioutil"
	"log"
	"time"
)

func main() {
	clientCert, err := tls.LoadX509KeyPair("cert/userclient-cert.pem", "cert/userclient-key.pem")
	if err != nil {
		log.Fatalf("Failed to load client certificate. %s.", err)
	}
	// Load the CA certificate
	trustedCert, err := ioutil.ReadFile("cert/server-ca-cert.pem")
	if err != nil {
		log.Fatalf("Failed to load trusted certificate. %s.", err)
	}
	// Put the CA certificate to certificate pool
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(trustedCert) {
		log.Fatalf("Failed to append trusted certificate to certificate pool. %s.", err)
	}

	// Create the TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
	}

	// Create a new TLS credentials based on the TLS configuration
	cred := credentials.NewTLS(tlsConfig)

	conn, err := grpc.Dial(":8010", grpc.WithTransportCredentials(cred))
	if err != nil {
		logrus.Fatalf("can not connect with server %v", err)
	}
	defer conn.Close()

	ctx := context.Background()
	client := proto.NewWorkerServiceClient(conn)

	res, err := client.StartJob(ctx, &proto.StartJobRequest{Cmd: "bash", Args: []string{"-c", "while true; do date; sleep 1; done"}})
	if err != nil {
		fmt.Errorf("failed to start job:%v", err)
	}
	jobID := res.GetID()
	fmt.Println(jobID)

	resp, err := client.GetOutputStream(ctx, &proto.GetStreamRequest{Id: jobID})
	if err != nil {
		log.Fatalf("could not start Job: %v", err)
	}
	done := make(chan bool)

	go func() {
		for {
			resp, err := resp.Recv()
			if err == io.EOF {
				done <- true //means stream is finished
				return
			}
			if err != nil {
				log.Fatalf("cannot receive %v", err)
			}
			log.Printf("Resp received: %s", resp.Result)
		}
	}()

	time.Sleep(20 * time.Second)

	_, err = client.StopJob(ctx, &proto.StopJobRequest{
		Id: jobID,
	})
	if err != nil {
		log.Fatalf("could not start Job: %v", err)
	}
}
