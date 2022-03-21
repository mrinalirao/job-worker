package server

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/mrinalirao/job-worker/proto"
	"github.com/mrinalirao/job-worker/store"
	"github.com/mrinalirao/job-worker/worker"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"net"
)

type Server struct {
	proto.UnimplementedWorkerServiceServer
	Worker       worker.Worker
	UserJobStore store.JobUserStore
}

func loadTLSCredentials() (credentials.TransportCredentials, error) {
	// Load the server certificate and its key
	//TODO: Pass certificates via config
	serverCert, err := tls.LoadX509KeyPair("cert/server-cert.pem", "cert/server-key.pem")
	if err != nil {
		return nil, fmt.Errorf("Failed to load server certificate and key. %w", err)
	}

	// Load the CA certificates
	trustedCert, err := ioutil.ReadFile("cert/client-ca-cert.pem")
	if err != nil {
		return nil, fmt.Errorf("Failed to load trusted certificate. %s.", err)
	}

	// add CA certificate to certificate pool
	certPool := x509.NewCertPool()
	if !certPool.AppendCertsFromPEM(trustedCert) {
		return nil, fmt.Errorf("Failed to append trusted certificate to certificate pool. %s.", err)
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
		MinVersion:   tls.VersionTLS13,
	}
	return credentials.NewTLS(config), nil
}

func createServer(cred credentials.TransportCredentials) (*grpc.Server, net.Listener, error) {
	// TODO: pass in server address from config file
	lis, err := net.Listen("tcp", ":8010")
	if err != nil {
		return nil, nil, err
	}
	userJobStore := store.NewJobStore()
	interceptor := NewInterceptor(userJobStore)
	grpcServer := grpc.NewServer(
		grpc.Creds(cred),
		grpc.UnaryInterceptor(interceptor.UnaryAuthInterceptor),
		grpc.StreamInterceptor(interceptor.StreamAuthInterceptor),
	)
	proto.RegisterWorkerServiceServer(grpcServer, &Server{
		Worker:       worker.NewWorker(),
		UserJobStore: userJobStore,
	})
	return grpcServer, lis, nil
}

func StartServer() error {
	cred, err := loadTLSCredentials()
	if err != nil {
		return err
	}
	serv, lis, err := createServer(cred)
	if err != nil {
		return err
	}
	logrus.Info("Listening server is created.")
	defer lis.Close()
	if err := serv.Serve(lis); err != nil {
		return err
	}
	return nil
}
