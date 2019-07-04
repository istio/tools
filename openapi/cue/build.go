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

// buildPlan specifies which .proto files end up in which OpenAPI .json files.
// The map maps from directory relative the root, to a grouping value for each
// file that needs to be generated.
var buildPlan = map[string][]grouping{
	"authentication/v1alpha1":     {{all: true}},
	"mcp/v1alpha1":                {{all: true}},
	"mesh/v1alpha1":               {{all: true}},
	"mixer/adapter/model/v1beta1": {{all: true}},
	"mixer/v1/config/client":      {{all: true}},
	"mixer/v1":                    {{all: true}},
	"networking/v1alpha3": {{
		oapiFilename: "istio.networking.destination.v1.json",
		protoFiles:   []string{"destination_rule.proto"},
	}, {
		oapiFilename: "istio.networking.envoy_filter.v1.json",
		protoFiles:   []string{"envoy_filter.proto"},
	}, {
		oapiFilename: "istio.networking.gateway.v1.json",
		protoFiles:   []string{"gateway.proto"},
	}, {
		oapiFilename: "istio.networking.service_entry.v1.json",
		protoFiles:   []string{"service_entry.proto"},
	}, {
		oapiFilename: "istio.networking.sidecar.v1.json",
		protoFiles:   []string{"sidecar.proto"},
	}},
	"policy/v1beta1": {{all: true}},
	"rbac/v1alpha1":  {{all: true}},
}

type grouping struct {
	dir string

	oapiFilename string // empty indicates the default name
	protoFiles   []string

	// pkg is used for custom packages, that only have a subset of the
	// proto files in them. For these CUE cannot derive the FQ name. But it will
	// always be the same package in question.
	pkg string

	// automatically add all files in this directory.
	all bool

	// derived automatically if unspecified.
	title   string
	version string
}

func completeBuildPlan(g map[string][]grouping, root string) error {
	// Walk over all .proto files in the root and add them to groupin entries
	// that requested all files in the directory to be added.
	err := filepath.Walk(root, func(path string, f os.FileInfo, _ error) (err error) {
		if !strings.HasSuffix(path, ".proto") {
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		dir, file := filepath.Split(rel)
		dir = filepath.Clean(dir)
		if len(buildPlan[dir]) == 0 || !buildPlan[dir][0].all {
			return nil
		}

		buildPlan[dir][0].protoFiles = append(buildPlan[dir][0].protoFiles, file)
		return nil
	})
	if err != nil {
		return err
	}

	// Complete version, titles, and file names
	for dir, all := range buildPlan {
		for i := range all {
			if err := (&all[i]).update(root, dir); err != nil {
				return err
			}
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

	return nil
}

func (g *grouping) update(root, dir string) error {
	g.dir = dir

	g.pkg = "istio." + strings.Replace(dir, "/", ".", -1)

	if g.oapiFilename == "" {
		g.oapiFilename = g.pkg + ".json"
	}

	g.version = filepath.Base(dir)

	if g.title == "" {
		for _, file := range g.protoFiles {
			if title, ok := findTitle(filepath.Join(root, dir, file)); ok {
				if g.title != "" && g.title != title {
					fmt.Printf("found two incompatible titles:\n\t%q, and\n\t%q", g.title, title)
				}
				g.title = title
			}
		}
	}
	return nil
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
