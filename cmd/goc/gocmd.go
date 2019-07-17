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
	"fmt"
	"os"
	"path"
	"runtime"
	"strings"
)

func goCmdPath() string {
	rt := runtime.GOROOT()
	return path.Join(rt, "bin", "go")
}

func goPath() string {
	return os.Getenv("GOPATH")
}

type gocmd struct {
	wd   string
	env  []string
	args []string
}

// currentGoCmd captures the current invocation of goc as a go invocation.
func currentGoCmd() (*gocmd, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("runGo: Getwd failure: %v", err)
	}

	return &gocmd{
		wd:   wd,
		args: os.Args,
		env:  os.Environ(),
	}, nil
}

func (g *gocmd) clone() *gocmd {
	env := make([]string, len(g.env))
	copy(env, g.env)

	args := make([]string, len(g.args))
	copy(args, g.args)

	return &gocmd{
		wd:   g.wd,
		env:  env,
		args: args,
	}
}

func (g *gocmd) isBuildCommand() bool {
	return len(g.args) > 1 && g.args[1] == "build"
}

func (g *gocmd) buildPackagePaths() []string {
	if g.args[1] != "build" {
		panic(fmt.Sprintf("gocmd.buildPackagePaths: command is not build: %v", g.args[1]))
	}
	var result []string

	var inFlag bool
	for _, a := range g.args[2:] {
		if inFlag {
			inFlag = false
			continue
		}

		if strings.HasPrefix(a, "-") {
			if !strings.Contains(a, "=") {
				inFlag = true
			}
			continue
		}

		result = append(result, a)
	}

	return result
}

func (g *gocmd) run() (int, error) {
	procArgs := &os.ProcAttr{
		Dir: g.wd,
		Env: g.env,
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	}

	p, err := os.StartProcess(goCmdPath(), os.Args, procArgs)
	if err != nil {
		return -1, fmt.Errorf("runGo: os.StartProcess: %v", err)
	}

	s, err := p.Wait()
	if err != nil {
		return -1, fmt.Errorf("runGo: os.Wait: %v", err)
	}

	return s.ExitCode(), nil
}
