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

const goldenDir = "testdata/golden/"

func TestOpenAPIGeneration(t *testing.T) {
	var testcases = []struct {
		name       string
		perPackage bool
		genOpts    string
		wantFiles  []string
	}{
		{
			name:       "Per Package Generation",
			perPackage: true,
			genOpts:    "",
			wantFiles:  []string{"testpkg.json", "testpkg2.json"},
		},
		{
			name:       "Single File Generation",
			perPackage: false,
			genOpts:    "single_file=true",
			wantFiles:  []string{"openapiv3.json"},
		},
		{
			name:       "Use $ref in the output",
			perPackage: false,
			genOpts:    "single_file=true,use_ref=true",
			wantFiles:  []string{"testRef/openapiv3.json"},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
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

			if tc.perPackage {
				for _, files := range packages {
					args := []string{"-Itestdata", "--openapi_out=" + tc.genOpts + ":" + tempDir}
					args = append(args, files...)
					protocOpenAPI(t, args)
				}
			} else {
				args := []string{"-Itestdata", "--openapi_out=" + tc.genOpts + ":" + tempDir}
				for _, files := range packages {
					args = append(args, files...)
				}
				protocOpenAPI(t, args)
			}

			// get the golden file and compare with the generated files.
			for _, file := range tc.wantFiles {
				wantPath := goldenDir + file
				// we are looking for the same file name in the generated path
				genPath := filepath.Join(tempDir, filepath.Base(wantPath))
				got, err := ioutil.ReadFile(genPath)
				if err != nil {
					if os.IsNotExist(err) {
						t.Fatalf("expected generated file %v does not exist: %v", genPath, err)
					} else {
						t.Errorf("error reading the generated file: %v", err)
					}
				}

				want, err := ioutil.ReadFile(wantPath)
				if err != nil {
					t.Errorf("error reading the golden file: %v", err)
				}

				if bytes.Equal(got, want) {
					return
				}

				cmd := exec.Command("diff", "-u", wantPath, genPath)
				out, _ := cmd.CombinedOutput()
				t.Errorf("golden file differs: %v\n%v", filepath.Base(wantPath), string(out))
			}
		})
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
