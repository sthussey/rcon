package internal

import (
	"fmt"
	"os"
    "runtime"
    "log"
    "golang.org/x/sys/unix"
)

type ProcessRunner interface {
	ValidateContext(p ProcessContext, errNoCap bool) []string
	ExecuteContext(p ProcessContext) bool
}

type LocalRunner struct {
}

func NewLocalRunner(cfg map[string]string) ProcessRunner {
	return LocalRunner{}
}

func (r LocalRunner) ValidateContext(p ProcessContext, errNoCap bool) []string {
	failures := make([]string, 0)

	msg := validateRunpath(p.RunPath)

	if msg != "" {
		failures = append(failures, msg)
	}

	msg = validateWorkingDir(p.WorkingDir)

	if msg != "" {
		failures = append(failures, msg)
	}

	return failures
}

func validateRunpath(p string) string {
	if p == "" {
		return "RunPath undefined."
	} else {
		rpstat, err := os.Stat(p)
		if err != nil {
			return fmt.Sprintf("Error stating RunPath %s: %v", p, err)
		} else {
			perms := rpstat.Mode().Perm()
			if (perms & 0111) == 0 {
				return fmt.Sprintf("RunPath %s found, but not executable.", p)
			}
		}
	}

	return ""
}

func validateWorkingDir(d string) string {
	if d != "" {
		wdstat, err := os.Stat(d)
		if err != nil {
			return fmt.Sprintf("Error stating WorkingDir %s: %v", d, err)
		} else {
			if !wdstat.IsDir() {
				return fmt.Sprintf("WorkingDir %s found, but not a directory.", d)
			}
		}
	}

	return ""
}

func (r LocalRunner) ExecuteContext(pctx ProcessContext) bool {
    doneChan := make(chan interface{})
    go executeInIsolatedThread(doneChan, pctx)
    <-doneChan
	return true
}

func executeInIsolatedThread(done chan<- interface{}, pctx ProcessContext) {
    runtime.LockOSThread()
    defer close(done)
    scratchPath, err := initMountNamespace()
    if err != nil {
        log.Printf("Error initializing mount namespace: %v", err)
        return
    }
    defer cleanMountNamespace(scratchPath)
	pa := os.ProcAttr{Dir: pctx.WorkingDir, Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}}
	p, err := os.StartProcess(pctx.RunPath, []string{pctx.RunPath}, &pa)
	if err != nil {
        log.Printf("Error starting tended process: %v", err)
		return
	}

	var ps *os.ProcessState
	ps, _ = p.Wait()

    if !ps.Success() {
        log.Printf("Tended process exitted unsuccesfully.")
    }

    // Deliberately leave thread locked so it isn't used by any other goroutines
    // and is instead kileld
    return
}


func initMountNamespace() (string, error) {
    unix.Unshare(unix.CLONE_NEWNS)
    tmpPath, err := os.MkdirTemp("", "")
    if err != nil {
        return "", fmt.Errorf("Error creating scratch mountpoint: %v", err)
    }
    err = unix.Mount("swap", tmpPath, "tmpfs", 0, "")
    if err != nil {
        return tmpPath, fmt.Errorf("Error mounting namespace scratch: %v", err)
    }
    return tmpPath, nil
}

func cleanMountNamespace(scratchPath string) {
    err := unix.Unmount(scratchPath, 0)
    if err != nil {
        log.Printf("Error unmounting scratch: %v", err)
        return
    }
    err = os.Remove(scratchPath)
    if err != nil {
        log.Printf("Error removing scratch mountpoint: %v", err)
        return
    }
    return
}
