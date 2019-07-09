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
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/emicklei/proto"
	"github.com/kr/pretty"
)

type plan map[string][]grouping

// buildPlan specifies which .proto files end up in which OpenAPI .json files.
// The map maps from directory relative the root, to a grouping value for each
// file that needs to be generated.
var buildPlan = plan{
	"authentication/v1alpha1":     {{all: true}},
	"mcp/v1alpha1":                {{all: true}},
	"mesh/v1alpha1":               {{all: true}},
	"mixer/adapter/model/v1beta1": {{all: true}},
	"mixer/v1/config/client":      {{all: true}},
	"mixer/v1":                    {{all: true}},
	"networking/v1alpha3":         {{perFile: true}},
	"policy/v1beta1":              {{all: true}},
	"rbac/v1alpha1":               {{all: true}},
}

type grouping struct {
	dir string

	oapiFilename string // empty indicates the default name
	protoFiles   []string

	// automatically add all files in this directory.
	all     bool
	perFile bool

	// derived automatically if unspecified.
	title   string
	version string
}

// fileFromDir computes the openapi json filename from the directory name.
// If filename is not "", it is assumed to be the proto filename in perFile
// mode.
func fileFromDir(dir, filename string) string {
	comps := strings.Split(dir, "/")
	if len(comps) == 0 {
		return "istio.json"
	}

	comps = append([]string{"istio"}, comps...)

	if filename == "" {
		return strings.Join(append(comps, "json"), ".")
	}

	filename = filename[:len(filename)-len(".proto")]
	version := len(comps) - 1
	return strings.Join(append(comps[:version], filename, comps[version], "json"), ".")
}

func completeBuildPlan(buildPlan plan, root string) (plan, error) {
	// Did the user override the plan with command lines?
	if *outdir != "" {
		*outdir = filepath.Clean(*outdir)
		buildPlan = plan{*outdir: {{
			all:     !*perFile,
			perFile: *perFile,
		}}}
	}

	// Walk over all .proto files in the root and add them to groupin entries
	// that requested all files in the directory to be added.
	err := filepath.Walk(root, func(path string, f os.FileInfo, _ error) (err error) {
		if !strings.HasSuffix(path, ".proto") {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		dir, file := filepath.Split(rel)
		dir = filepath.Clean(dir)
		switch {
		case len(buildPlan[dir]) == 0:
			return nil
		case buildPlan[dir][0].perFile:
			if len(buildPlan[dir][0].protoFiles) > 0 {
				buildPlan[dir] = append(buildPlan[dir], grouping{
					protoFiles: []string{file},
					perFile:    true,
				})
				break
			}
			fallthrough
		case buildPlan[dir][0].all:
			buildPlan[dir][0].protoFiles = append(buildPlan[dir][0].protoFiles, file)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	// Complete version, titles, and file names
	for dir, all := range buildPlan {
		for i := range all {
			(&all[i]).update(root, dir)
		}
	}

	if *verbose {
		pretty.Print(buildPlan)
		fmt.Println()
	}

	// Validate
	for dir, all := range buildPlan {
		for _, g := range all {
			if g.title == "" {
				g.title = "NO TITLE"
				fmt.Printf("No $description set for package %q in any .proto file\n", dir)
			}
		}
	}

	return buildPlan, nil
}

func (g *grouping) update(root, dir string) {
	g.dir = dir

	if g.oapiFilename == "" {
		filename := ""
		if g.perFile && len(g.protoFiles) > 0 {
			filename = g.protoFiles[0]
		}
		g.oapiFilename = fileFromDir(dir, filename)
	}

	g.version = filepath.Base(dir)

	if g.title == "" {
		for _, file := range g.protoFiles {
			if title, ok := findTitle(filepath.Join(root, dir, file)); ok {
				if g.title != "" && g.title != title {
					fmt.Printf("found two incompatible titles for %s:\n\t%q, and\n\t%q\n", g.oapiFilename, g.title, title)
				}
				g.title = title
			}
		}
	}
}

var descRe = regexp.MustCompile(`\$description: (.*)`)

func findTitle(filename string) (title string, ok bool) {
	for _, d := range protoElems(filename) {
		switch x := d.(type) {
		case *proto.Comment:
			for _, str := range x.Lines {
				m := descRe.FindStringSubmatch(str)
				if m != nil {
					return m[1], true
				}
			}
		case *proto.Package:
			return "", false
		}
	}
	return "", false
}

func protoElems(filename string) []proto.Visitee {
	reader, err := os.Open(filename)
	if err != nil {
		log.Fatal("Could not find proto file:", err)
	}
	defer reader.Close()

	parser := proto.NewParser(reader)
	def, err := parser.Parse()
	if err != nil {
		log.Fatal("Error parsing proto:", err)
	}
	return def.Elements
}
