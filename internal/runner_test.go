package internal

import (
	"fmt"
	"testing"
)

func TestRenderEnvNoPropagation(t *testing.T) {
	t.Parallel()
	var testEnv = make(map[string]string, 1)

	var testKey = "foo"
	var testValue = "bar"

	testEnv[testKey] = testValue

	testResult := renderEnvironment(testEnv, nil)

	if len(testResult) != 1 {
		t.Errorf("Render result had wrong number of entries: %d", len(testResult))
	}

	if testResult[0] != fmt.Sprintf("%s=%s", testKey, testValue) {
		t.Errorf("Render value '%s' does not match expected '%s=%s'", testResult[0], testKey, testValue)
	}
}

func TestRenderEnvBasicPropagation(t *testing.T) {
	t.Parallel()
	var existingEnv = []string{"baz=bar"}
	var testEnv = make(map[string]string, 1)

	var testKey = "foo"
	var testValue = "bar"

	testEnv[testKey] = testValue

	testResult := renderEnvironment(testEnv, existingEnv)

	if len(testResult) != 2 {
		t.Errorf("Render result had wrong number of entries: %d", len(testResult))
	}

	expectedEnv := append(existingEnv, fmt.Sprintf("%s=%s", testKey, testValue))

	for _, e := range testResult {
		if !contains(expectedEnv, e) {
			t.Errorf("Rendered result contains unexpected value '%s'", e)
		}
	}
}

func TestRenderEnvDeletePropagation(t *testing.T) {
	t.Parallel()
	var existingEnv = []string{"baz=bar"}
	var testEnv = make(map[string]string, 1)

	var testKey = "foo"
	var testValue = "bar"
	var testDeleteKey = "baz"

	testEnv[testKey] = testValue
	testEnv[testDeleteKey] = ""

	testResult := renderEnvironment(testEnv, existingEnv)

	if len(testResult) != 1 {
		t.Errorf("Render result had wrong number of entries: %d", len(testResult))
	}

	expectedEnv := []string{fmt.Sprintf("%s=%s", testKey, testValue)}

	for _, e := range testResult {
		if !contains(expectedEnv, e) {
			t.Errorf("Rendered result contains unexpected value '%s'", e)
		}
	}
}

func contains(array []string, element string) bool {
	for _, e := range array {
		if e == element {
			return true
		}
	}
	return false
}
