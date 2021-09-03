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
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"golang.org/x/tools/go/packages"
)

// instrumenter is used to create instrumented version of the code.
type instrumenter struct {
	// the new GOPATH, where instrumented code will be placed.
	goPath string

	initialPaths []string
}

func newInstrumenter() *instrumenter {
	return &instrumenter{}
}

// instrument the code, if needed, and return a new gocmd that will generate the binary with the instrumented code when
// invoked.
func (i *instrumenter) instrument(g *gocmd) (*gocmd, error) {
	if !g.isBuildCommand() {
		return g, nil
	}

	i.initialPaths = g.buildPackagePaths()

	cfg := packages.Config{Mode: packages.NeedFiles | packages.NeedSyntax}
	pkgs, err := packages.Load(&cfg, i.initialPaths...)
	if err != nil {
		return nil, err
	}

	if packages.PrintErrors(pkgs) > 0 {
		fmt.Printf("goc: There seems to be an error in compilation. Skipping instrumentation...\n")
		return g, nil
	}

	// Calculate new gopath
	p, err := ioutil.TempDir(os.TempDir(), "goc")
	if err != nil {
		return nil, err
	}
	i.goPath = p

	// TODO: This needs to be repo agnostic
	if err = os.MkdirAll(path.Join(i.goPath, "src", "istio.io/istio"), os.ModePerm); err != nil {
		return nil, err
	}

	// TODO: This needs to be repo agnostic
	err = os.Symlink(path.Join(goPath(), "src", "istio.io/istio/vendor"), path.Join(i.goPath, "src", "istio.io/istio/vendor"))
	if err != nil {
		return nil, err
	}

	if err = os.MkdirAll(path.Join(i.goPath, "src"), os.ModePerm); err != nil {
		return nil, err
	}

	for _, info := range pkgs {
		if !isInInstrumentationScope(info) {
			continue
		}

		if err = i.instrumentPackage(info); err != nil {
			return nil, err
		}
	}

	g = g.clone()
	g.wd = path.Join(i.goPath, "src", "istio.io/istio") // TODO: This needs to be repo agnostic.
	var env []string
	for _, e := range g.env {
		if !strings.HasPrefix(e, "GOPATH") {
			env = append(env, e)
		}
	}
	env = append(env, fmt.Sprintf("GOPATH=%s", i.goPath))
	g.env = env
	return g, nil
}

func isInInstrumentationScope(info *packages.Package) bool {
	// TODO: This needs to be repo agnostic
	// TODO: A better way to calculate scope. (Environment variables?)
	return strings.HasPrefix(info.PkgPath, "istio.io")
}

func (i *instrumenter) instrumentPackage(info *packages.Package) error {
	oldPkgPath := path.Join(goPath(), "src", info.PkgPath)
	newPkgPath := path.Join(i.goPath, "src", info.PkgPath)

	files, err := ioutil.ReadDir(oldPkgPath)
	if err != nil {
		return err
	}

	contextVars := make(map[string]string)
	for _, f := range files {
		if f.IsDir() {
			continue
		}
		if !strings.HasSuffix(f.Name(), ".go") {
			continue
		}

		if strings.HasSuffix(f.Name(), "_test.go") {
			continue
		}

		oldFilePath := path.Join(oldPkgPath, f.Name())
		newFilePath := path.Join(newPkgPath, f.Name())

		varName := generateVarName(oldFilePath)
		if err := instrumentFile(varName, oldFilePath, newFilePath); err != nil {
			return err
		}

		context := path.Join(info.PkgPath, f.Name())
		contextVars[context] = varName
	}
	if err := generateRegistrationFile(info, newPkgPath, contextVars); err != nil {
		return err
	}

	return nil
}

func generateRegistrationFile(info *packages.Package, pkgPath string, contextVars map[string]string) error {
	regFile := path.Join(pkgPath, "codecovreg.go")

	rendered, err := generateRegistrationCode(info.Name, contextVars)
	if err != nil {
		return err
	}

	if err = ioutil.WriteFile(regFile, []byte(rendered), os.ModePerm); err != nil {
		return err
	}

	return nil
}

func generateVarName(filePath string) string {
	r := sha256.Sum256([]byte(filePath))
	in := make([]byte, len(r))
	for i := 0; i < len(r); i++ {
		in[i] = r[i]
	}
	return "codeCov" + hex.EncodeToString(in)
}

func instrumentFile(varName, oldFilePath, newFilePath string) error {
	procArgs := &os.ProcAttr{
		Files: []*os.File{
			os.Stdin,
			os.Stdout,
			os.Stderr,
		},
	}

	args := []string{
		"go",
		"tool",
		"cover",
		"-var",
		varName,
		"-mode",
		"atomic",
		"-o",
		newFilePath,
		oldFilePath,
	}

	if err := os.MkdirAll(path.Dir(newFilePath), os.ModePerm); err != nil {
		return err
	}

	p, err := os.StartProcess(goCmdPath(), args, procArgs)
	if err != nil {
		return fmt.Errorf("instrumentFile: os.StartProcess: %v", err)
	}

	_, err = p.Wait()
	if err != nil {
		return fmt.Errorf("instrumentFile: os.Wait: %v", err)
	}

	return nil
}

// Close implements io.Closer
func (i *instrumenter) Close() error {
	if i.goPath != "" {
		return os.RemoveAll(i.goPath)
	}
	return nil
}
