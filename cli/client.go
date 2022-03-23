package cli

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"github.com/mrinalirao/job-worker/proto"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"os"
)

type ClientConfig struct {
	CertFile string
	KeyFile  string
	CAFile   string
}

func NewClientConfig(caPath string, keyPath string, certPath string) ClientConfig {
	return ClientConfig{
		CAFile:   caPath,
		KeyFile:  keyPath,
		CertFile: certPath,
	}
}

func loadTLSCredentials(cc ClientConfig) (credentials.TransportCredentials, error) {
	clientCert, err := tls.LoadX509KeyPair(cc.CertFile, cc.KeyFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load server certificate and key. %w", err)
	}
	// Load the CA certificate
	trustedCert, err := os.ReadFile(cc.CAFile)
	if err != nil {
		return nil, fmt.Errorf("failed to load trusted certificate. %s.", err)
	}
	// Put the CA certificate to certificate pool
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(trustedCert) {
		return nil, errors.New("failed to append certificate pem")
	}

	// Create the TLS configuration
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{clientCert},
		RootCAs:      certPool,
		MinVersion:   tls.VersionTLS13,
		MaxVersion:   tls.VersionTLS13,
	}
	return credentials.NewTLS(tlsConfig), nil
}

func NewWorkerClient(config ClientConfig) (proto.WorkerServiceClient, error) {
	dialOptions := grpc.WithInsecure()
	if config.CAFile != "" && config.CertFile != "" && config.KeyFile != "" {
		transportCredentials, err := loadTLSCredentials(config)
		if err != nil {
			return nil, err
		}
		dialOptions = grpc.WithTransportCredentials(transportCredentials)
	}
	conn, err := grpc.Dial(":8010", dialOptions)
	if err != nil {
		logrus.Fatalf("can not connect with server %v", err)
	}

	return proto.NewWorkerServiceClient(conn), nil
}

func StartJobHandler(ctx context.Context, client proto.WorkerServiceClient, params *Params) {
	resp, err := client.StartJob(ctx, &proto.StartJobRequest{
		Cmd:  params.CommandName,
		Args: params.Arguments,
	})
	if err != nil {
		logrus.Fatalf("failed to start job %v", err)
	}
	logrus.Infof("started JobID: %v", resp.GetID())
}

func StopJobHandler(ctx context.Context, client proto.WorkerServiceClient, params *Params) {
	jobID := params.JobID
	_, err := client.StopJob(ctx, &proto.StopJobRequest{
		Id: jobID,
	})
	if err != nil {
		logrus.Fatalf("failed to stop job %v", err)
	}
	logrus.Infof("stopped Job: %v", params.JobID)
}

func GetJobStatusHandler(ctx context.Context, client proto.WorkerServiceClient, params *Params) {
	jobID := params.JobID
	resp, err := client.GetJobStatus(ctx, &proto.GetStatusRequest{
		Id: jobID,
	})
	if err != nil {
		logrus.Fatalf("failed to fetch job status: %v", err)
	}
	logrus.Infof(" jobID: %v, status: %v, exitCode: %v ", jobID, resp.GetStatus(), resp.Exitcode)
}

func GetJobOutputHandler(ctx context.Context, client proto.WorkerServiceClient, params *Params) {
	jobID := params.JobID
	resp, err := client.GetOutputStream(ctx, &proto.GetStreamRequest{Id: jobID})
	if err != nil {
		log.Fatalf("could not get output of Job: %v", err)
	}

	for {
		select {
		case <-resp.Context().Done():
			logrus.Fatalf(resp.Context().Err().Error())
			return
		default:
			l, err := resp.Recv()
			if err != nil {
				if err == io.EOF {
					return
				}
				logrus.Fatalf("failed to receive log line on stream: %v", err)
				return
			}
			if b := l.GetResult(); len(b) > 0 {
				logrus.Infof("%s", b)
			}
		}
	}
}
