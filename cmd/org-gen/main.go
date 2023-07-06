// Copyright Istio Authors.
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
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"

	"istio.io/tools/pkg/orggen"
)

var (
	input       = flag.String("input", "org", "folder to read config from")
	output      = flag.String("output", "", "file to write config to")
	write       = flag.Bool("write-to-github", false, "if set, will actually commit results to GitHub")
	githubToken = flag.String("github-token", "", "the filepath to a GitHub token to use with --write-to-github")
)

func main() {
	flag.Parse()
	out := *output
	if out == "" {
		var err error
		outf, err := os.CreateTemp("", "org")
		if err != nil {
			exit(err)
		}
		out = outf.Name()
	}
	log.Printf("Reading from %v, writing to %v", *input, out)
	cfg, err := orggen.ReadConfig(*input)
	if err != nil {
		exit(err)
	}
	org := orggen.ConvertConfig(cfg)
	if err := orggen.WriteConfig(org, out); err != nil {
		exit(err)
	}
	if *write {
		// We could import peribolos but the dependency is huge, just exec it...
		c := exec.Command(
			"peribolos",
			"--fix-org",
			"--fix-org-members",
			"--fix-teams",
			"--fix-team-members",
			"--fix-team-repos",
			"--config-path", out,
			"--github-token-path", *githubToken,
			"--confirm",
		)
		c.Stderr = os.Stderr
		c.Stdout = os.Stdout
		if err := c.Run(); err != nil {
			exit(err)
		}
	}
}

func exit(err error) {
	fmt.Fprintf(os.Stderr, "failed to generate org: %v\n", err)
	os.Exit(-1)
}
