// Copyright 2020 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"testing"
)

func TestReadAsJson(t *testing.T) {
	executablePath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get executable path: %s\n", err.Error())
	}

	_, err = readYAMLAsJSON(fmt.Sprintf("%s/%s", executablePath, "/testdata/features.yaml"))
	if err != nil {
		t.Fatalf("Failed to read document as JSON: %s\n", err.Error())
	}
}

func TestParse(t *testing.T) {
	executablePath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get executable path: %s\n", err.Error())
	}

	cmd := exec.Command("go", "run", "validate.go", "--documentPath",
		fmt.Sprintf("%s/testdata/features.yaml", executablePath),
		"--schemaPath", fmt.Sprintf("%s/testdata/features_schema.json", executablePath))
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Failed to validate: %s stdout: %s\n", err.Error(), cmdOutput)
	}
}

func TestParseBroken(t *testing.T) {
	executablePath, err := os.Getwd()
	if err != nil {
		t.Fatalf("Unable to get executable path: %s\n", err.Error())
	}

	cmd := exec.Command("go", "run", "validate.go",
		"--documentPath", fmt.Sprintf("%s/testdata/features_broken.yaml", executablePath),
		"--schemaPath", fmt.Sprintf("%s/testdata/features_schema.json", executablePath))
	cmdOutput := &bytes.Buffer{}
	cmd.Stdout = cmdOutput
	err = cmd.Run()
	if err == nil {
		t.Fatalf("Unexpectedly validated without error. Stdout: %s\n", cmdOutput)
	}
}
