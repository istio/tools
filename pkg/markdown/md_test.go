// Copyright 2018 Istio Authors. All Rights Reserved.
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

package markdown

import (
	"fmt"
	"os"
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name string
	}{
		{
			name: "AnalysisMessageWeakSchema",
		},
		{
			name: "links",
		},
		{
			name: "telemetry",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in, err := readInput(tc.name)
			assert.NoError(t, err)
			out, err := readOutput(tc.name)
			assert.NoError(t, err)

			got := Run(in)
			assert.Equal(t, string(out), string(got))
		})
	}
}

func readInput(name string) ([]byte, error) {
	return readFile(fmt.Sprintf("%s.input", name))
}

func readOutput(name string) ([]byte, error) {
	return readFile(fmt.Sprintf("%s.output", name))
}

func readFile(f string) ([]byte, error) {
	return os.ReadFile(path.Join("testdata", f))
}
