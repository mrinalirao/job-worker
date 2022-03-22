package server

import (
	"context"
	"errors"
	"github.com/mrinalirao/job-worker/proto"
	"github.com/mrinalirao/job-worker/store"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
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
	jobID := jobIdFromRequest(req)
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

func (r *recvWrapper) Context() context.Context {
	return r.ctx
}

func (r *recvWrapper) RecvMsg(m interface{}) error {
	if err := r.ServerStream.RecvMsg(m); err != nil {
		logrus.Errorf("failed to intercept stream: %v", err)
		return err
	}
	if req, ok := m.(*proto.GetStreamRequest); ok {
		jobID := req.GetId()
		newCtx, err := r.authorize(r.ctx, "/proto.WorkerService/GetOutputStream", jobID)
		if err != nil {
			return err
		}
		r.ctx = newCtx
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
		return ctx, errors.New("error to read peer information")
	}
	// reads user tls information
	tlsInfo, ok := peer.AuthInfo.(credentials.TLSInfo)
	if !ok {
		return ctx, errors.New("error to get auth information")
	}
	// get user roles
	certs := tlsInfo.State.VerifiedChains
	if len(certs) == 0 || len(certs[0]) == 0 {
		return ctx, errors.New("missing certificate chain")
	}

	// find user roles from certificate extensions
	var roles []string
	for _, ext := range certs[0][0].Extensions {
		if ext.Id.Equal(oidRole) {
			roles = ParseRoles(string(ext.Value))
			break
		}
	}
	// check user has access to execute a specific method
	if !HasAccess(method, roles) {
		return ctx, errors.New("unauthorized, user does not have privileges")
	}

	// Get subject common name
	userName := certs[0][0].Subject.CommonName

	if jobID != "" && !contains("admin", roles) {
		u, err := i.jobUserStore.GetUser(jobID)
		if err != nil {
			return ctx, errors.New("failed to verify user access to job")
		}
		if u != userName {
			return ctx, errors.New("no does not have access to this job")
		}
	}
	return context.WithValue(ctx, userKey{}, &User{
		Name: userName,
	}), nil
}

// jobIdFromRequest returns the jobID from the request
func jobIdFromRequest(req interface{}) string {
	switch r := req.(type) {
	case *proto.StopJobRequest:
		return r.GetId()
	case *proto.GetStatusRequest:
		return r.GetId()
	default:
	}
	return ""
}
