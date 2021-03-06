package internal

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
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

	msg = validateEnvironment(p.Environment)

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

func validateEnvironment(varMap map[string]string) string {
	for k := range varMap {
		if strings.Contains(k, "=") {
			return "Environment variable keys cannot contain '='"
		}
	}
	return ""
}

func renderEnvironment(varMap map[string]string, prop []string) []string {
	var procEnv []string = make([]string, 0)
	for _, v := range prop {
		// Here check if the declaritive environment contains
		// a propagated variable. If do, do not propagate. This
		// allows declaritive intent to unset a variable
		splitVal := strings.SplitN(v, "=", 2)
		_, ok := varMap[splitVal[0]]
		if !ok {
			procEnv = append(procEnv, v)
		}
	}
	for k, v := range varMap {
		if v != "" {
			procEnv = append(procEnv, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return procEnv
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
	setupMounts(scratchPath, pctx.Files)
	defer cleanMountNamespace(scratchPath, pctx.Files)
	pa := os.ProcAttr{
		Dir:   pctx.WorkingDir,
		Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
		Env: renderEnvironment(pctx.Environment,
			func() []string {
				if pctx.PropEnv {
					return os.Environ()
				}
				return nil
			}())}

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
}
