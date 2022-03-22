package worker

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"
)

const cgroupPath = "/sys/fs/cgroup"

var testmode = false

func addCgroupLimit(jobID string, pid int, maxCPUBandwidth *uint32, maxMemoryUsageBytes *uint64, maxRBPS *uint64, maxWBPS *uint64, maxRIOPS *uint64, maxWIOPS *uint64) error {
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
		maxWIOPS != nil {
		path := filepath.Join(cgPath, "io.max")
		buf := &bytes.Buffer{}
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
