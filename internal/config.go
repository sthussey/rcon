package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type RunConfig struct {
	ConfigPath string
	RunContext ProcessContext
}

type ProcessContext struct {
	RunPath    string `json:"runPath"`
	WorkingDir string `json:"workingDir"`
}

func NewRunConfig(p string) RunConfig {
	var rc RunConfig
	if p != "" {
		rc = RunConfig{ConfigPath: p}
	} else {
		rc = RunConfig{}
	}
    selectParser(&rc)
	return rc
}

func selectParser(rc *RunConfig) error {
	if rc.ConfigPath != "" {
        fmt.Printf("Loading config from %s\n", rc.ConfigPath);
		json, err := ioutil.ReadFile(rc.ConfigPath)
		if err != nil {
			return fmt.Errorf("Error opening config file %s: %w", rc.ConfigPath, err)
		}
		var pctx *ProcessContext
		pctx, err = jsonParseConfig(json)
		if err != nil {
			return fmt.Errorf("Error parsing config file %s: %w", rc.ConfigPath, err)
		} else {
            fmt.Printf("Loading config: \n%v\n", pctx);
        }
		rc.RunContext = *pctx
	} else {
		envmap := make(map[string]string)
		for _, v := range os.Environ() {
			splitv := strings.SplitN(v, "=", 2)
			envmap[splitv[0]] = splitv[1]
		}
		var pctx *ProcessContext
		pctx, err := envParseConfig(envmap)
		if err != nil {
			return fmt.Errorf("Error parsing environment config: %w", err)
		}
		rc.RunContext = *pctx
	}
	return nil
}

func jsonParseConfig(src []byte) (*ProcessContext, error) {
	pctx := ProcessContext{}
	err := json.Unmarshal(src, &pctx)
	if err != nil {
		return &pctx, fmt.Errorf("JSON parsing error: %w", err)
	}
	return &pctx, nil
}

func envParseConfig(src map[string]string) (*ProcessContext, error) {
	pctx := ProcessContext{}
	pctx.RunPath = src["RCON_RUNPATH"]
	pctx.WorkingDir = src["RCON_WORKINGDIR"]
	return &pctx, nil
}
