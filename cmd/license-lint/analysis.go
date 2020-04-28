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
	"fmt"
	"io/ioutil"

	"gopkg.in/src-d/go-license-detector.v2/licensedb"
	"gopkg.in/src-d/go-license-detector.v2/licensedb/filer"
)

// analysisResult describes a license.
type analysisResult struct {
	licenseName          string
	confidence           float32
	similarLicense       string
	similarityConfidence float32
}

// filerImpl implements filer.Filer to return the license text directly
// from the github.RepositoryLicense structure.
type filerImpl struct {
	License string
}

func (f *filerImpl) ReadFile(name string) ([]byte, error) {
	return ioutil.ReadFile(name)
}

func (f *filerImpl) ReadDir(dir string) ([]filer.File, error) {
	// We only support root
	if dir != "" {
		return nil, nil
	}

	return []filer.File{{Name: f.License}}, nil
}

func (f *filerImpl) Close() {}

func analyzeLicense(path string) (analysisResult, error) {
	res, err := licensedb.Detect(&filerImpl{License: path})
	if err == licensedb.ErrNoLicenseFound {
		return analysisResult{}, nil
	}
	if err != nil {
		return analysisResult{}, fmt.Errorf("failed to detect license %v: %v", path, err)
	}
	// Find the highest matching license
	var confidence float32
	licenseName := ""
	for id, v := range res {
		if v > confidence {
			confidence = v
			licenseName = id
		}
	}

	similarLicense := ""
	var similarityConfidence float32 = 0.0

	if confidence < 0.85 {
		// Not enough confidence
		similarLicense = licenseName
		similarityConfidence = confidence
		licenseName = ""
		confidence = 0.0
	}

	return analysisResult{
		licenseName:          licenseName,
		confidence:           confidence,
		similarLicense:       similarLicense,
		similarityConfidence: similarityConfidence,
	}, nil
}
