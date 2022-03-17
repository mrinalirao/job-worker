package worker

import (
	"context"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestWorker_Start(t *testing.T) {
	w := NewWorker()
	jobID, err := w.Start("echo", []string{"foo"})
	assert.Nil(t, err)
	assert.NotEmpty(t, jobID)
}

func TestWorker_StartNonExistingCommand(t *testing.T) {
	w := NewWorker()
	jobID, err := w.Start("xyz", []string{"foo"})
	assert.NotEmpty(t, jobID)
	assert.NotNil(t, err)
}

func TestWorker_Stop(t *testing.T) {
	w := NewWorker()
	jobID, err := w.Start("sleep", []string{"4"})
	assert.Nil(t, err)
	err = w.Stop(jobID)
	assert.Nil(t, err)
}

func TestWorker_StopNonExistingJob(t *testing.T) {
	randomJobID, _ := uuid.NewRandom()
	w := NewWorker()
	err := w.Stop(randomJobID.String())
	assert.NotNil(t, err)
}

func TestWorker_GetStatus(t *testing.T) {
	w := NewWorker()
	jobID, err := w.Start("sleep", []string{"1"})
	assert.NotEmpty(t, jobID)
	assert.NoError(t, err)

	stat, err := w.GetStatus(jobID)
	assert.Nil(t, err)
	assert.Equal(t, Running, stat.JobStatus)

	err = w.Stop(jobID)
	assert.Nil(t, err)

	stat, err = w.GetStatus(jobID)
	assert.Nil(t, err)
	assert.Equal(t, Stopped, stat.JobStatus)
}

func TestWorker_StreamExistingProcess(t *testing.T) {
	w := NewWorker()
	jobID, err := w.Start("bash", []string{"-c", "while true; do date; sleep 1; done"})
	assert.Nil(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	logchan, err := w.GetOutput(ctx, jobID)
	assert.Nil(t, err)

	assert.NotNil(t, <-logchan)
	cancel()

	err = w.Stop(jobID)
	assert.NoError(t, err)
}

func TestWorker_StreamNonExistingProcess(t *testing.T) {
	w := NewWorker()
	randomJobID, _ := uuid.NewRandom()
	logchan, err := w.GetOutput(context.Background(), randomJobID.String())
	assert.Error(t, err)
	assert.Nil(t, logchan)
}
