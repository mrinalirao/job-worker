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
const bufferSize = 1024

type logger struct {
	logStore string
	sync.RWMutex
}

func newLogger() *logger {
	return &logger{
		logStore: os.TempDir(),
	}
}

// CreateFile creates and returns a os.File. If the file can't be created an error will be returned.
func (l *logger) CreateFile(fileName string) (*os.File, error) {
	path := filepath.Join(l.logStore, fmt.Sprintf("%s.log", fileName))
	return os.Create(path)
}

// RemoveFile deletes the named file under the log store.
func (l *logger) RemoveFile(fileName string) error {
	path := filepath.Join(l.logStore, fmt.Sprintf("%s.log", fileName))
	return os.Remove(path)
}

// TailReader waits until new data is written to file instead of returning io.EOF
func (l *logger) TailReader(ctx context.Context, jobID string, doneCh chan struct{}) (<-chan string, error) {
	path := filepath.Join(l.logStore, fmt.Sprintf("%s.log", jobID))
	file, err := os.OpenFile(path, os.O_RDONLY, 0644)
	if err != nil {
		return nil, err
	}

	//watcher to track file changes internally
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		logrus.Errorf("failed to create a watcher: %w", err)
		return nil, err
	}

	err = watcher.Add(path)
	if err != nil {
		logrus.Errorf("failed to watch file: %w", err)
		return nil, err
	}
	outputChan := make(chan string)
	go func() {
		defer func() {
			if err := file.Close(); err != nil {
				logrus.Errorf("fail to close the log file: %w", err)
			}
			close(outputChan)
			if err := watcher.Close(); err != nil {
				logrus.Errorf("failed to close file watcher: %w", err)
			}
		}()
		// reads from the begin
		if err := l.sendOutputTail(ctx, outputChan, file); err != nil {
			logrus.Errorf("failed to stream output: %w", err)
			return
		}

		// reads file changes
		for {
			select {
			case <-ctx.Done():
				logrus.Error(ctx.Err())
				return
			case <-doneCh:
				if err := l.sendOutputTail(ctx, outputChan, file); err != nil {
					logrus.Errorf("failed to stream output: %w", err)
					return
				}
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				logrus.Debugf("event: %v", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					if err := l.sendOutputTail(ctx, outputChan, file); err != nil {
						logrus.Errorf("failed to stream output: %w", err)
						return
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				logrus.Error("error:", err)
				return
			}
		}
	}()
	return outputChan, nil
}

func (l *logger) sendOutputTail(ctx context.Context, outputChan chan<- string, file *os.File) error {
	for {
		// TODO: pass configs from a config file
		buf := make([]byte, bufferSize)
		n, err := file.Read(buf)
		if err != nil {
			if n > 0 {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
					outputChan <- string(buf[:n])
				}
			}
			if err != io.EOF {
				return fmt.Errorf("failed to read file: %w", err)
			}
			return nil
		}
		select {
		case outputChan <- string(buf[:n]):
		case <-ctx.Done():
			return errors.New(fmt.Sprintf("output stream cancelled: %w", ctx.Err()))
		}
	}
}
