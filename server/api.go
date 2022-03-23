package server

import (
	"context"
	"github.com/mrinalirao/job-worker/proto"
	"github.com/mrinalirao/job-worker/worker"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *Server) StartJob(ctx context.Context, r *proto.StartJobRequest) (*proto.StartJobResponse, error) {
	logFields := logrus.Fields{
		"Action": "StartJob",
	}
	log := logrus.WithFields(logFields)
	user, ok := UserFromContext(ctx)
	if !ok || user.Name == "" {
		return nil, status.Errorf(codes.Internal, "failed to verify user")
	}

	jobID, err := s.Worker.Start(r.Cmd, r.Args)
	if err != nil {
		log.WithError(err).Error("failed to start job")
		//Note: we intentionally do not expose the errors to the user as the errors might contain internal implementation details.
		// eg: Unable to remove file, in a prod system we will use different Error types rather than passing error strings and map them to error codes in gRPC.
		return nil, status.Errorf(codes.Internal, "failed to start job")
	}

	if err := s.UserJobStore.SetJobUser(jobID, user.Name); err != nil {
		log.WithError(err).Error("failed to start job")
		return nil, status.Errorf(codes.Internal, "failed to save job for user")
	}

	res := proto.StartJobResponse{
		ID: jobID,
	}
	return &res, nil
}

func (s *Server) StopJob(ctx context.Context, in *proto.StopJobRequest) (*proto.StopJobResponse, error) {
	jobID := in.GetId()
	logFields := logrus.Fields{
		"JobID":  jobID,
		"Action": "StopJob",
	}
	err := s.Worker.Stop(jobID)
	if err != nil {
		logrus.WithFields(logFields).Error(err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to stop job: %v", jobID)
	}
	return &proto.StopJobResponse{}, nil
}

func (s *Server) GetJobStatus(ctx context.Context, in *proto.GetStatusRequest) (*proto.GetStatusResponse, error) {
	jobID := in.GetId()
	logFields := logrus.Fields{
		"JobID":  jobID,
		"Action": "GetStatus",
	}
	stat, err := s.Worker.GetStatus(jobID)
	if err != nil {
		logrus.WithFields(logFields).Error(err)
		return nil, status.Errorf(codes.InvalidArgument, "failed to fetch status for job: %v", jobID)
	}
	var jobStatus proto.Status
	switch stat.JobStatus {
	case worker.Running:
		jobStatus = proto.Status_RUNNING
	case worker.Finished:
		jobStatus = proto.Status_FINISHED
	case worker.Stopped:
		jobStatus = proto.Status_STOPPED
	default:
		logrus.WithFields(logFields).Errorf("job with invalid status: %v", stat.JobStatus)
		return nil, status.Errorf(codes.InvalidArgument, "job: %v has invalid status", jobID)
	}
	return &proto.GetStatusResponse{
		Status:   jobStatus,
		Exitcode: int32(stat.ExitCode),
	}, nil
}

func (s *Server) GetOutputStream(r *proto.GetStreamRequest, stream proto.WorkerService_GetOutputStreamServer) error {
	jobID := r.GetId()
	logFields := logrus.Fields{
		"JobID":  jobID,
		"Action": "GetOutputStream",
	}
	logchan, err := s.Worker.GetOutput(stream.Context(), jobID)
	if err != nil {
		logrus.WithFields(logFields).Error(err)
		return status.Errorf(codes.Internal, "failed to get stream output of job: %v", jobID)
	}
	for {
		select {
		case <-stream.Context().Done():
			return stream.Context().Err()
		case log, ok := <-logchan:
			if !ok {
				return nil
			}
			if err := stream.SendMsg(&proto.GetStreamResponse{Result: []byte(log)}); err != nil {
				logrus.WithFields(logFields).Error(err)
				return status.Errorf(codes.Internal, "failed to get stream output of job: %v", jobID)
			}
		}
	}
}
