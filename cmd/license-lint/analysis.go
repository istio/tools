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
	"os/exec"
	"strings"
)

// analysisResult describes a license.
type analysisResult struct {
	licenseName string
	exactMatch  bool
	confidence  string
}

func analyzeLicense(path string) (analysisResult, error) {
	// run external licensee tool to analyze a specific license
	b, err := exec.Command("licensee", "detect", path).Output()
	if err != nil {
		return analysisResult{}, err
	}
	out := string(b)
	lines := strings.Split(out, "\n")

	// extract core analysis result
	licenseName := getValue(lines, "License:")
	confidence := getValue(lines, "  Confidence:")

	if licenseName == "NOASSERTION" {
		// For NOASSERTION license type, it means we are below the match threshold. Still grab the closest match and output
		// confidence value.
		licenseName = "UNKNOWN"
		confidence = ""
		for _, l := range lines {
			if strings.Contains(l, " similarity:") {
				fs := strings.Fields(l)
				licenseName = fs[0]
				confidence = fs[2]
			}
		}
	}

	return analysisResult{
		licenseName: licenseName,
		confidence:  confidence,
		exactMatch:  strings.Contains(out, "Licensee::Matchers::Exact"),
	}, nil
}

func getValue(lines []string, match string) string {
	for _, l := range lines {
		if strings.Contains(l, match) {
			return strings.TrimSpace(strings.TrimPrefix(l, match))
		}
	}
	return ""
}
