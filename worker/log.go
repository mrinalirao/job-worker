package worker

import (
	"context"
	"errors"
	"fmt"
	"github.com/fsnotify/fsnotify"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"path/filepath"
	"sync"
)

//The buffer size 1024 is just chosen randomly, performance is only affected when the process writes huge amount of data to the file
const BUFFER_SIZE = 1024

type Logger struct {
	LogStore    string
	LogListener map[string]chan string
	sync.RWMutex
}

func NewLogger() *Logger {
	return &Logger{
		LogStore:    os.TempDir(),
		LogListener: make(map[string]chan string),
	}
}

// Subscribe subscribes to the job
func (l *Logger) Subscribe(userID string) chan string {
	c := make(chan string)
	l.Lock()
	l.LogListener[userID] = c
	l.Unlock()
	return c
}

// Unsubscribe unsubscribes from the job
func (l *Logger) Unsubscribe(userID string) error {
	l.RLock()
	c, ok := l.LogListener[userID]
	l.RUnlock()
	if !ok {
		return errors.New("user is not a listener")
	}

	l.Lock()
	delete(l.LogListener, userID)
	close(c)
	l.Unlock()

	return nil
}

// CreateFile creates and returns a os.File. If the file can't be created an error will be returned.
func (l *Logger) CreateFile(fileName string) (*os.File, error) {
	path := filepath.Join(l.LogStore, fmt.Sprintf("%s.log", fileName))
	return os.Create(path)
}

// RemoveFile deletes the named file under the log store.
func (l *Logger) RemoveFile(fileName string) error {
	path := filepath.Join(l.LogStore, fmt.Sprintf("%s.log", fileName))
	return os.Remove(path)
}

// TailReader waits until new data is written to file instead of returning io.EOF
func (l *Logger) TailReader(ctx context.Context, jobID string, userID string) (chan string, error) {
	outputChan := l.Subscribe(userID)
	path := filepath.Join(l.LogStore, fmt.Sprintf("%s.log", jobID))
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}
	go func() {
		defer func() {
			l.Unsubscribe(userID)
			if err := file.Close(); err != nil {
				logrus.Errorf("fail to close the log file: %v", err)
			}
		}()
		// reads from the begin
		if err := l.streamOutput(ctx, outputChan, file); err != nil && err != io.EOF {
			logrus.Errorf("failed to read the file: %v", err)
			return
		}

		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			logrus.Errorf("failed to create a watcher: %v", err)
			return
		}
		defer watcher.Close()
		err = watcher.Add(path)
		if err != nil {
			logrus.Errorf("failed to watch file: %v", err)
			return
		}

		// reads file changes
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				logrus.Debugf("event: %v", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					if err := l.streamOutput(ctx, outputChan, file); err != nil && err != io.EOF {
						logrus.Errorf("fail to read the log file: %v", err)
						return
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Error("error:", err)
			}
		}
	}()
	return outputChan, nil
}

func (l *Logger) streamOutput(ctx context.Context, outputChan chan string, file *os.File) error {
	for {
		// TODO: pass configs from a config file
		buf := make([]byte, BUFFER_SIZE)
		n, err := file.Read(buf)
		if err != nil {
			return err
		}
		select {
		case outputChan <- string(buf[:n]):
		case <-ctx.Done():
			return errors.New("output stream cancelled")
		}
	}
}
