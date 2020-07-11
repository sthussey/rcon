package internal

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type RunConfig struct {
	configPath string
	runContext ProcessContext
}

type ProcessContext struct {
	RunPath    string `json:"runPath"`
	WorkingDir string `json:"workingDir"`
}

func NewRunConfig(p string) RunConfig {
	var rc RunConfig
	if p != "" {
		rc = RunConfig{configPath: p}
	} else {
		rc = RunConfig{}
	}
	return rc
}

func (rc RunConfig) Run() error {
	err := rc.loadConfig()
	if err != nil {
		return fmt.Errorf("Error parsing config: %w", err)
	}
	return nil
}

func (rc RunConfig) ConfigPath() string {
	return rc.configPath
}

func (rc *RunConfig) selectParser() error {
	if rc.ConfigPath() != "" {
		json, err := ioutil.ReadFile(rc.ConfigPath())
		if err != nil {
			return fmt.Errorf("Error opening config file %s: %w", rc.ConfigPath(), err)
		}
		var pctx *ProcessContext
		pctx, err = jsonParseConfig(json)
		if err != nil {
			return fmt.Errorf("Error parsing config file %s: %w", rc.ConfigPath(), err)
		}
		rc.runContext = *pctx
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
		rc.runContext = *pctx
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

func (rc *RunConfig) loadConfig() error {
	rc.selectParser()
	return nil
}
