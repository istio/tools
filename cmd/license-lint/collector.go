// Copyright Istio Authors
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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// information about a single go module
type moduleInfo struct {
	moduleName string
	path       string
	licenses   []*licenseInfo
}

// information about a single license file
type licenseInfo struct {
	module   *moduleInfo
	path     string
	text     string
	analysis analysisResult
}

// a go module, as returned by `go list -m`
type module struct {
	Path      string    `json:"Path"`
	Version   string    `json:"Version"`
	Replace   *module   `json:"Replace"`
	Time      time.Time `json:"Time"`
	Main      bool      `json:"Main"`
	Indirect  bool      `json:"Indirect"`
	Dir       string    `json:"Dir"`
	GoMod     string    `json:"GoMod"`
	GoVersion string    `json:"GoVersion"`
}

func getLicenses() ([]*moduleInfo, error) {
	// find all the modules this repo depends on
	mods, err := getDependentModules()
	if err != nil {
		return nil, err
	}

	var result []*moduleInfo
	for _, m := range mods {

		if m.Dir == "" {
			return nil, fmt.Errorf("couldn't find content of module %s (did you forget to do `go mod download`?)", m.Path)
		}

		// find all the license files contained in the module
		licenseFiles, err := findLicenseFiles(m.Dir)
		if err != nil {
			return nil, err
		}

		mi := &moduleInfo{
			moduleName: m.Path,
			path:       m.Dir,
		}

		for _, f := range licenseFiles {
			// read each license file
			text, err := ioutil.ReadFile(f)
			if err != nil {
				return nil, fmt.Errorf("unable to read license file %s: %v", f, err)
			}

			li := licenseInfo{
				module: mi,
				path:   f,
				text:   string(text),
			}

			// analyze each license file
			li.analysis, err = analyzeLicense(f)
			if err != nil {
				return nil, err
			}

			mi.licenses = append(mi.licenses, &li)
		}
		sort.Slice(mi.licenses, func(i, j int) bool {
			return strings.Compare(mi.licenses[i].path, mi.licenses[j].path) < 0
		})

		result = append(result, mi)
	}

	sort.Slice(result, func(i, j int) bool {
		return strings.Compare(result[i].moduleName, result[j].moduleName) < 0
	})

	return result, nil
}

func getDependentModules() ([]module, error) {
	cmd := exec.Command("go", "list", "-m", "-json", "all")

	// Turn on Go module support
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "GO111MODULE=on")
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return nil, fmt.Errorf("%v: %v", err, stderr)
	}

	// Unmarshal json output
	var modules []module

	// Need to add `,`` between arrays in json output and add []
	fixed := "[\n" + strings.Replace(stdout.String(), "}\n{", "},\n{", -1) + "\n]"

	err = json.Unmarshal([]byte(fixed), &modules)
	if err != nil {
		return nil, fmt.Errorf("unable to decode module list: %v", err)
	}

	return modules, nil
}

// the set of license files we recognize
var supportedLicenseFilenames = map[string]struct{}{
	"LICENSE": {},
	// nolint: misspell
	"LICENCE":      {},
	"LICENSE.TXT":  {},
	"LICENCE.TXT":  {},
	"LICENSE.MD":   {},
	"LICENCE.MD":   {},
	"LICENSE.CODE": {},
	"LICENCE.CODE": {},
	"COPYING":      {},
}

// find all license files in the given directory tree
func findLicenseFiles(path string) ([]string, error) {
	var result []string
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if info == nil {
			return fmt.Errorf("unable to get information on %s: %v", path, err)
		}

		if !info.IsDir() {
			name := strings.ToUpper(info.Name())
			if _, ok := supportedLicenseFilenames[name]; ok {
				result = append(result, path)
				return nil
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}
