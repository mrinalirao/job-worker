package server

import (
	"context"
	"fmt"
	"github.com/mrinalirao/job-worker/proto"
	"github.com/mrinalirao/job-worker/store"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

type interceptor struct {
	jobUserStore store.JobUserStore
}

func NewInterceptor(store store.JobUserStore) *interceptor {
	return &interceptor{
		jobUserStore: store,
	}
}

// UnaryAuthInterceptor intercept unary calls to authorize the user
// It checks user role using certification extension oid 1.2.840.10070.8.1, it also checks if user has access to the requested resource
func (i *interceptor) UnaryAuthInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	jobID := interceptRequest(info.FullMethod, req)
	newCtx, err := i.authorize(ctx, info.FullMethod, jobID)
	if err != nil {
		return nil, status.Error(codes.PermissionDenied, err.Error())
	}
	return handler(newCtx, req)
}

type recvWrapper struct {
	grpc.ServerStream
	ctx context.Context
	*interceptor
}

func (r *recvWrapper) RecvMsg(m interface{}) error {
	if err := r.ServerStream.RecvMsg(m); err != nil {
		logrus.Errorf("failed to intercept stream: %v", err)
		return err
	}
	if req, ok := m.(*proto.GetStreamRequest); ok {
		jobID := req.GetId()
		newCtx, err := r.authorize(r.ctx, "/proto.WorkerService/GetJobStream", jobID)
		if err != nil {
			return err
		}
		md, _ := metadata.FromIncomingContext(newCtx)
		r.ctx = metadata.NewIncomingContext(newCtx, md)
	}
	return nil
}

// StreamAuthInterceptor intercept stream calls to authorize the user
// It checks the user role using certification extension oid 1.2.840.10070.8.1, it also checks if user has access to the requested resource
func (i *interceptor) StreamAuthInterceptor(srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
	wrapper := &recvWrapper{stream, stream.Context(), i}
	return handler(srv, wrapper)
}

// authorize verifies the user information given by certificate
// against the mapped roles for a specific method and also checks if user has access to requested job
func (i *interceptor) authorize(ctx context.Context, method string, jobID string) (context.Context, error) {
	// reads the peer information from context
	peer, ok := peer.FromContext(ctx)
	if !ok {
		return ctx, fmt.Errorf("error to read peer information")
	}
	// reads user tls information
	tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return ctx, fmt.Errorf("error to get auth information")
	}
	// get user roles
	certs := tlsInfo.State.VerifiedChains
	if len(certs) == 0 || len(certs[0]) == 0 {
		return ctx, fmt.Errorf("missing certificate chain")
	}

	// find user roles from certificate extensions
	var roles []string
	for _, ext := range certs[0][0].Extensions {
		if oid := OidToString(ext.Id); IsOidRole(oid) {
			roles = ParseRoles(string(ext.Value))
			break
		}
	}
	// check user has access to execute a specific method
	if !HasAccess(method, roles) {
		return ctx, fmt.Errorf("unauthorized, user does not have privileges")
	}

	// Get subject common name
	userName := certs[0][0].Subject.CommonName

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ctx, fmt.Errorf("")
	}
	// inject username into context
	md.Append("username", userName)
	c := metadata.NewIncomingContext(ctx, md)

	if jobID != "" && userName != "adminuser" {
		u, err := i.jobUserStore.GetUser(jobID)
		if err != nil {
			return ctx, fmt.Errorf("failed to verify user access to job")
		}
		if u != userName {
			return c, fmt.Errorf("no does not have access to this job")
		}
	}
	return c, nil
}

// Intercepts the request to get the jobID
func interceptRequest(method string, req interface{}) string {
	switch req.(type) {
	case *proto.StartJobRequest:
		return ""
	case *proto.StopJobRequest:
		if req, ok := req.(*proto.StopJobRequest); ok {
			return req.GetId()
		}
	case *proto.GetStatusRequest:
		if req, ok := req.(*proto.GetStatusRequest); ok {
			return req.GetId()
		}
	default:
		return ""
	}
	return ""
}
