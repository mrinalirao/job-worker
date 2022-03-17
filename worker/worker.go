package worker

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type StatusEnum int

const (
	Running StatusEnum = iota
	Stopped
	Finished
)

//Worker defines the operations to manage Jobs.
type Worker interface {
	Start(cmdName string, args []string) (string, error)
	Stop(jobID string) error
	GetStatus(jobID string) (Status, error)
	GetOutput(ctx context.Context, jobID string) (<-chan string, error)
}

// job represents a Linux process scheduled by the Worker.
type job struct {
	id       uuid.UUID
	cmdName  string
	args     []string
	status   StatusEnum
	exitCode int
	cmd      *exec.Cmd
	doneChan chan struct{} // closed when done running

}

type worker struct {
	// log is responsible to handle the output of a job
	log  *logger
	jobs map[string]*job
	sync.RWMutex
}

// Status of the job.
type Status struct {
	JobStatus StatusEnum
	ExitCode  int
}

// NewWorker creates a new Worker instance.
func NewWorker() Worker {
	return &worker{
		jobs: make(map[string]*job),
		log:  newLogger(),
	}
}

// Starts a linux process and assigns a uuid to the underlying process.
// A log file with the JobID name is created to capture the output of the running process
func (w *worker) Start(cmdName string, args []string) (string, error) {
	jobID := uuid.New()
	fileName := jobID.String()
	logfile, err := w.log.CreateFile(fileName)
	if err != nil {
		return "", err
	}
	cmd := exec.Command(cmdName, args...)
	cmd.Stdout = logfile
	cmd.Stderr = logfile

	if err := cmd.Start(); err != nil {
		if err := w.log.RemoveFile(fileName); err != nil {
			logrus.Errorf("Unable to remove file, err: %v", err)
		}
		return jobID.String(), err
	}

	job := &job{
		id:       jobID,
		cmdName:  cmdName,
		args:     args,
		cmd:      cmd,
		status:   Running,
		doneChan: make(chan struct{}),
	}

	w.Lock()
	//TODO: Add pid to cgroup

	w.jobs[jobID.String()] = job
	w.Unlock()
	go w.run(job)
	return jobID.String(), nil
}

func (w *worker) run(j *job) {
	defer close(j.doneChan)
	//Wait for the cmd to be finished or killed
	if err := j.cmd.Wait(); err != nil {
		logrus.WithFields(logrus.Fields{
			"Job ID": j.id,
			"Name":   j.cmdName,
			"Args":   j.args}).Errorf("execution failed: %v", err)
	}
	w.Lock()
	j.exitCode = j.cmd.ProcessState.ExitCode()
	if j.status != Stopped {
		j.status = Finished
	}
	w.Unlock()
}

// Stops the underlying linux job with the given JobID
func (w *worker) Stop(jobID string) error {
	w.Lock()
	defer w.Unlock()
	job, found := w.jobs[jobID]
	if !found {
		return fmt.Errorf("job %v not found", jobID)
	}

	select {
	case <-job.doneChan:
		return nil
	default:
		// NOTE: This potentially is in race condition with Wait call in the run goroutine started by Start,
		// so we check for ErrProcessDone even though we acquired the lock.
		switch err := job.cmd.Process.Signal(syscall.SIGKILL); err {
		case os.ErrProcessDone:
			return nil
		default:
			job.status = Stopped
			return err
		}
	}
}

// GetStatus returns the status of the job with the given JobID.
func (w *worker) GetStatus(jobID string) (Status, error) {
	w.RLock()
	job, found := w.jobs[jobID]
	if !found {
		return Status{}, fmt.Errorf("job %v not found", jobID)
	}
	// return a copy of status to avoid data races
	stat, exitCode := job.status, job.exitCode
	w.RUnlock()

	return Status{stat, exitCode}, nil
}

// GetOutput reads from the log file. If the context is canceled the channel will
// be closed and the tailing will be stopped.
func (w *worker) GetOutput(ctx context.Context, jobID string) (<-chan string, error) {
	w.RLock()
	job, found := w.jobs[jobID]
	w.RUnlock()
	if !found {
		return nil, fmt.Errorf("job %v not found", jobID)
	}
	return w.log.TailReader(ctx, job.id.String(), job.doneChan)
}
