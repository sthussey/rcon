package internal

import (
	"fmt"
	"os"
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
	pa := os.ProcAttr{Dir: pctx.WorkingDir, Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}}
	p, err := os.StartProcess(pctx.RunPath, []string{pctx.RunPath}, &pa)
	if err != nil {
		return false
	}

	var ps *os.ProcessState
	ps, _ = p.Wait()
	return ps.Success()
}
