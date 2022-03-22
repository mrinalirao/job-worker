package worker

import (
	"bytes"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

const cgroupPath = "/sys/fs/cgroup"

var testmode = false

func addCgroupLimit(jobID string, pid int, maxCPUBandwidth *uint32, maxMemoryUsageBytes *uint64, maxRBPS *uint64, maxWBPS *uint64, maxRIOPS *uint64, maxWIOPS *uint64, majorDevice *uint64, minorDevice *uint64) error {
	if testmode {
		return nil
	}
	cgPath := filepath.Join(cgroupPath, jobID)
	if err := os.MkdirAll(cgPath, 0700); err != nil {
		return fmt.Errorf("failed to create cgroup: %w", err)
	}

	if err := os.WriteFile(filepath.Join(cgPath, "cgroup.subtree_control"), []byte("+cpu +memory +io"), syscall.O_WRONLY); err != nil {
		return fmt.Errorf("failed to add controllers to cgroup: %w", err)
	}

	if maxCPUBandwidth != nil {
		if err := os.WriteFile(filepath.Join(cgPath, "cpu.max"), []byte(strconv.FormatUint(uint64(*maxCPUBandwidth), 10)), syscall.O_WRONLY); err != nil {
			return fmt.Errorf("failed to write 'cpu.max': %w", err)
		}
	} else {
		// set default CPU limit
		if err := os.WriteFile(filepath.Join(cgPath, "cpu.max"), []byte("600000 1000000"), syscall.O_WRONLY); err != nil {
			return fmt.Errorf("failed to write 'cpu.max': %w", err)
		}
	}

	if maxMemoryUsageBytes != nil {
		if err := os.WriteFile(filepath.Join(cgPath, "memory.max"), []byte(strconv.FormatUint(*maxMemoryUsageBytes, 10)), syscall.O_WRONLY); err != nil {
			return fmt.Errorf("failed to write 'memory.max': %w", err)
		}
	} else {
		// set default Memory limit to 50MB
		if err := os.WriteFile(filepath.Join(cgPath, "memory.max"), []byte("50000000"), syscall.O_WRONLY); err != nil {
			return fmt.Errorf("failed to write 'cpu.max': %w", err)
		}
	}

	if maxRBPS != nil ||
		maxWBPS != nil ||
		maxRIOPS != nil ||
		maxWIOPS != nil ||
		majorDevice != nil ||
		minorDevice != nil {
		path := filepath.Join(cgPath, "io.max")
		buf := &bytes.Buffer{}
		buf.WriteString(fmt.Sprintf("%d:%d", majorDevice, minorDevice))
		if maxRBPS != nil {
			fmt.Fprintf(buf, " rbps=%d", *maxRBPS)
		}
		if maxWBPS != nil {
			fmt.Fprintf(buf, " wbps=%d", *maxWBPS)
		}
		if maxRIOPS != nil {
			fmt.Fprintf(buf, " riops=%d", *maxRIOPS)
		}
		if maxWIOPS != nil {
			fmt.Fprintf(buf, " wiops=%d", *maxWIOPS)
		}
		if err := os.WriteFile(path, buf.Bytes(), syscall.O_WRONLY); err != nil {
			return fmt.Errorf("failed to write 'io.max': %w", err)
		}
	} else {
		if err := os.WriteFile(filepath.Join(cgPath, "io.max"), []byte("254:0 wbps=1048576"), syscall.O_WRONLY); err != nil {
			return fmt.Errorf("failed to write 'io.max': %w", err)
		}
	}

	if err := os.WriteFile(filepath.Join(cgPath, "cgroup.procs"), []byte(strconv.Itoa(pid)), syscall.O_WRONLY); err != nil {
		return fmt.Errorf("failed to add process to cgroup: %w", err)
	}

	return nil
}

func rmdir(path string) error {
	err := unix.Rmdir(path)
	if err == nil || err == unix.ENOENT { // unix errors are bare
		return nil
	}
	return &os.PathError{Op: "rmdir", Path: path, Err: err}
}

// RemovePath aims to remove cgroup path. It does so recursively,
// by removing any subdirectories (sub-cgroups) first.
func RemovePath(jobID string) error {
	if testmode {
		return nil
	}
	path := filepath.Join(cgroupPath, jobID)
	// try the fast path first
	if err := rmdir(path); err == nil {
		return nil
	}

	infos, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return err
	}
	for _, info := range infos {
		if info.IsDir() {
			// We should remove subcgroups dir first
			if err = RemovePath(filepath.Join(path, info.Name())); err != nil {
				break
			}
		}
	}
	if err == nil {
		err = os.Remove(path)
	}
	return err
}
