package internal

import (
	"fmt"
	"testing"
)

func TestJsonLoad(t *testing.T) {
	runpath := getContextValues("RunPath")
	workingdir := getContextValues("WorkingDir")
	sample := []byte(fmt.Sprintf("{\"runpath\": \"%s\", \"workingdir\": \"%s\"}", runpath, workingdir))
	result, err := jsonParseConfig(sample)
	if err != nil {
		t.Errorf("Valid JSON Load Failed: %v", err)
	}
	validateContext(t, result)
}

func TestEnvLoad(t *testing.T) {
	sample := make(map[string]string)
	sample["RCON_RUNPATH"] = getContextValues("RunPath")
	sample["RCON_WORKINGDIR"] = "/tmp"

	result, err := envParseConfig(sample)
	if err != nil {
		t.Errorf("Valid environment load failed: %v", err)
	}
	validateContext(t, result)
}

/* Reusable source for building test contexts
   k is a key to get a value for
*/
func getContextValues(k string) string {
	switch k {
	case "RunPath":
		return "/bin/false"
	case "WorkingDir":
		return "/tmp"
	default:
		return ""
	}
}

func validateContext(t *testing.T, pctx *ProcessContext) {
	if pctx.RunPath != getContextValues("RunPath") {
		t.Errorf("Context has wrong RunPath - Found %s, expected %s", pctx.RunPath, getContextValues("RunPath"))
	}
	if pctx.WorkingDir != getContextValues("WorkingDir") {
		t.Errorf("Context has wrong WorkingDir - Found %s, expected %s", pctx.WorkingDir, getContextValues("WorkingDir"))
	}
}
