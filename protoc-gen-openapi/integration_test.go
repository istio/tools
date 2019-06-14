// Copyright 2019 Istio Authors
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
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestOpenAPIGeneration(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "openapi-temp")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// we assume that the package name is the same as the name of the folder containing the proto files.
	packages := make(map[string][]string)
	err = filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".proto") {
			dir := filepath.Dir(path)
			packages[dir] = append(packages[dir], path)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}

	for _, files := range packages {
		args := []string{"-Itestdata", "--openapi_out=mode=true:" + tempDir}
		args = append(args, files...)
		protocOpenAPI(t, args)
	}

	// find the golden files in the test directories and compare with the generated files.
	err = filepath.Walk("testdata", func(path string, info os.FileInfo, err error) error {
		if strings.HasSuffix(path, ".json") {
			filename := info.Name()
			genPath := filepath.Join(tempDir, filename)
			got, err := ioutil.ReadFile(genPath)
			if err != nil {
				t.Errorf("error reading the generated file: %v", err)
				return nil
			}

			want, err := ioutil.ReadFile(path)
			if err != nil {
				t.Errorf("error reading the golden file: %v", err)
			}

			if bytes.Equal(got, want) {
				return nil
			}

			cmd := exec.Command("diff", "-u", path, genPath)
			out, _ := cmd.CombinedOutput()
			t.Errorf("golden file differs: %v\n%v", path, string(out))
			return nil
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func protocOpenAPI(t *testing.T, args []string) {
	cmd := exec.Command("protoc", "--plugin=protoc-gen-openapi="+os.Args[0])
	cmd.Args = append(cmd.Args, args...)
	cmd.Env = append(os.Environ(), "RUN_AS_PROTOC_GEN_OPENAPI=1")
	out, err := cmd.CombinedOutput()
	if len(out) > 0 || err != nil {
		t.Log("RUNNING: ", strings.Join(cmd.Args, " "))
	}
	if len(out) > 0 {
		t.Log(string(out))
	}
	if err != nil {
		t.Fatalf("protoc: %v", err)
	}
}

func init() {
	// when "RUN_AS_PROTOC_GEN_OPENAPI" is set, we use the protoc-gen-openapi directly
	// for the test scenarios.
	if os.Getenv("RUN_AS_PROTOC_GEN_OPENAPI") != "" {
		main()
		os.Exit(0)
	}
}
