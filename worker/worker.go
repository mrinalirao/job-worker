package worker

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"os/exec"
	"sync"
	"syscall"
)

const (
	JSTATUS_START   = "running"
	JSTATUS_END     = "finished"
	JSTATUS_STOPPED = "stopped"
)

//Worker defines the operations to manage Jobs.
type Worker interface {
	Start(cmdName string, args []string) (string, error)
	Stop(jobID string) error
	GetStatus(jobID string) (Status, error)
	GetOutput(ctx context.Context, jobID string, userID string) (chan string, error)
}

// Job represents a Linux process scheduled by the Worker.
type Job struct {
	ID       uuid.UUID
	CmdName  string
	Args     []string
	Status   string
	ExitCode int
	Cmd      *exec.Cmd
}

type worker struct {
	// log is responsible to handle the output of a job
	log  *Logger
	jobs map[string]*Job
	sync.RWMutex
}

// Status of the job.
type Status struct {
	JobStatus string
	ExitCode  int
}

// NewWorker creates a new Worker instance.
func NewWorker() Worker {
	return &worker{
		jobs: make(map[string]*Job),
		log:  NewLogger(),
	}
}

// Starts a linux process and assigns a uuid to the underlying process.
// A log file with the JobID name is created to capture the output of the running process
func (w *worker) Start(cmdName string, args []string) (string, error) {
	jobID := uuid.New()
	fileName := jobID.String()
	logfile, err := w.log.CreateFile(fileName)
	if err != nil {
		return jobID.String(), err
	}
	cmd := exec.Command(cmdName, args...)
	cmd.Stdout = logfile
	cmd.Stderr = logfile

	if err := cmd.Start(); err != nil {
		if err := w.log.RemoveFile(fileName); err != nil {
			logrus.Errorf("Unable to remove file")
			return jobID.String(), err
		}
		return jobID.String(), err
	}

	job := &Job{
		ID:      jobID,
		CmdName: cmdName,
		Args:    args,
		Cmd:     cmd,
	}

	w.Lock()
	//TODO: Add pid to cgroup
	job.Status = JSTATUS_START
	w.jobs[jobID.String()] = job
	w.Unlock()
	go w.run(job)
	return jobID.String(), nil
}

func (w *worker) run(j *Job) {
	//Wait for the cmd to be finished or killed
	if err := j.Cmd.Wait(); err != nil {
		logrus.WithFields(logrus.Fields{
			"Job ID": j.ID,
			"Name":   j.CmdName,
			"Args":   j.Args}).Errorf("execution failed: %v", err)
	}
	w.Lock()
	j.ExitCode = j.Cmd.ProcessState.ExitCode()
	if j.Status != JSTATUS_STOPPED {
		j.Status = JSTATUS_END
	}
	w.Unlock()
}

// Stops the underlying linux job with the given JobID
func (w *worker) Stop(jobID string) error {
	w.RLock()
	job, found := w.jobs[jobID]
	w.RUnlock()
	if !found {
		return fmt.Errorf("job %v not found", jobID)
	}
	w.Lock()
	if job.Status != JSTATUS_END {
		job.Status = JSTATUS_STOPPED
		job.Cmd.Process.Signal(syscall.SIGKILL)
	}
	w.Unlock()
	return nil
}

// GetStatus returns the status of the job with the given JobID.
func (w *worker) GetStatus(jobID string) (Status, error) {
	w.RLock()
	job, found := w.jobs[jobID]
	w.RUnlock()
	if !found {
		return Status{}, fmt.Errorf("job %v not found", jobID)
	}
	return Status{job.Status, job.ExitCode}, nil
}

// GetOutput reads from the log file. If the context is canceled the channel will
// be closed and the tailing will be stopped.
func (w *worker) GetOutput(ctx context.Context, jobID string, userID string) (chan string, error) {
	w.RLock()
	job, found := w.jobs[jobID]
	w.RUnlock()
	if !found {
		return nil, fmt.Errorf("job %v not found", jobID)
	}
	return w.log.TailReader(ctx, job.ID.String(), userID)
}
