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

type Plan map[string][]Grouping

// buildPlan specifies which .proto files end up in which OpenAPI .json files.
// The map maps from directory relative the root, to a grouping value for each
// file that needs to be generated.
var buildPlan = Plan{
	"authentication/v1alpha1":     {{All: true}},
	"mcp/v1alpha1":                {{All: true}},
	"mesh/v1alpha1":               {{All: true}},
	"mixer/adapter/model/v1beta1": {{All: true}},
	"mixer/v1/config/client":      {{All: true}},
	"mixer/v1":                    {{All: true}},
	"networking/v1alpha3":         {{PerFile: true}},
	"policy/v1beta1":              {{All: true}},
	"rbac/v1alpha1":               {{All: true}},
}

type Grouping struct {
	dir string

	OapiFilename string // empty indicates the default name
	ProtoFiles   []string

	// automatically add all files in this directory.
	All     bool
	PerFile bool

	// derived automatically if unspecified.
	Title   string
	Version string
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

func completeBuildPlan(buildPlan Plan, root string) (Plan, error) {
	// Did the user override the Plan with command lines?
	if *outdir != "" {
		*outdir = filepath.Clean(*outdir)
		buildPlan = Plan{*outdir: {{
			All:     !*perFile,
			PerFile: *perFile,
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
		case buildPlan[dir][0].PerFile:
			if len(buildPlan[dir][0].ProtoFiles) > 0 {
				buildPlan[dir] = append(buildPlan[dir], Grouping{
					ProtoFiles: []string{file},
					PerFile:    true,
				})
				break
			}
			fallthrough
		case buildPlan[dir][0].All:
			buildPlan[dir][0].ProtoFiles = append(buildPlan[dir][0].ProtoFiles, file)
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
			if g.Title == "" {
				g.Title = "NO TITLE"
				fmt.Printf("No $description set for package %q in any .proto file\n", dir)
			}
		}
	}

	return buildPlan, nil
}

func (g *Grouping) update(root, dir string) {
	g.dir = dir

	if g.OapiFilename == "" {
		filename := ""
		if g.PerFile && len(g.ProtoFiles) > 0 {
			filename = g.ProtoFiles[0]
		}
		g.OapiFilename = fileFromDir(dir, filename)
	}

	g.Version = filepath.Base(dir)

	if g.Title == "" {
		for _, file := range g.ProtoFiles {
			if title, ok := findTitle(filepath.Join(root, dir, file)); ok {
				if g.Title != "" && g.Title != title {
					fmt.Printf("found two incompatible titles for %s:\n\t%q, and\n\t%q\n", g.OapiFilename, g.Title, title)
				}
				g.Title = title
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
