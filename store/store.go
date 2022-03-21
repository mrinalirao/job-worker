package store

import (
	"fmt"
	"github.com/sirupsen/logrus"
	"sync"
)

// jobUserStore keeps a map of jobIDs to their userIDs, this is required prevent unauthorized access to jobs
type jobUserStore struct {
	jobUserMap map[string]string
	sync.RWMutex
}

type JobUserStore interface {
	SetJobUser(jobID string, userID string) error
	GetUser(jobID string) (string, error)
}

func NewJobStore() JobUserStore {
	return &jobUserStore{
		jobUserMap: make(map[string]string),
	}
}

func (j *jobUserStore) SetJobUser(jobID string, userID string) error {
	logFields := logrus.Fields{
		"Action": "SetUser",
		"JobID":  jobID,
		"UserID": userID,
	}
	j.Lock()
	defer j.Unlock()
	if v, ok := j.jobUserMap[jobID]; ok {
		logrus.WithFields(logFields).Errorf("job already exists with user: %s", v)
		if v != userID {
			return fmt.Errorf("failed to set job:%s for user: %s", jobID, userID)
		}
		return nil
	}
	j.jobUserMap[jobID] = userID
	return nil
}

func (j *jobUserStore) GetUser(jobID string) (string, error) {
	j.RLock()
	defer j.RUnlock()
	v, ok := j.jobUserMap[jobID]
	if !ok {
		return "", fmt.Errorf("job does not exist: %v", jobID)
	}
	return v, nil
}
