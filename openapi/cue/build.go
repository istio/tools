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

	"cuelang.org/go/cue"
	"cuelang.org/go/encoding/openapi"
	"cuelang.org/go/encoding/yaml"
	"github.com/emicklei/proto"
	"github.com/kr/pretty"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// A Config defines the OpenAPI to generate and their properties.
type Config struct {
	// Module is the Go or CUE modules for which to generated OpenAPI
	// definitions.
	Module string

	cwd string // the current working directory

	// The generator configuration.
	Openapi *openapi.Generator

	// Directories is a list of files to generate per directory.
	Directories map[string][]Grouping

	// Information about the output of an aggregate OpenAPI file.
	All *Grouping

	// Crd is the configuration for CRD generation.
	Crd *CrdGen
}

const (
	perFile  = "perFile"
	allFiles = "all"
)

// Grouping defines the source and settings for a single file.
//
// See doc.cue for more information on these fields.
type Grouping struct {
	dir string

	OapiFilename string // empty indicates the default name

	// Mode defines the set of files to include by default:
	//   manual   user defines ProtoFiles
	//   all      all proto files in this directory are automatically added
	//   perFile  a single file is generated for each proto file in the directory
	Mode string

	// ProtoFiles defines the list of proto files to include as the bases
	// of the generated file. The paths are relative the the directory.
	ProtoFiles []string

	// derived automatically if unspecified.
	Title   string
	Version string
}

// CrdGen defines the output of the CRD file.
type CrdGen struct {
	Dir string // empty indicates the default directory.

	Filename string // empty indicates the default prefix.

	// Mapping of CRD name and its output configuration.
	CrdConfigs map[string]CrdConfig
}

// CrdConfig defines the details about each CRD to be generated.
type CrdConfig struct {
	// the name of the schema to use for this CRD.
	SchemaName string

	// the base of the CRD.
	Metadata metav1.ObjectMeta
	Spec     apiext.CustomResourceDefinitionSpec
}

func loadConfig(filename string) (c *Config, err error) {
	r := &cue.Runtime{}

	f, err := docCueBytes()
	if err != nil {
		return nil, err
	}
	inst, err := r.Compile("doc.cue", f)
	if err != nil {
		log.Fatal(err)
	}

	var cfg *cue.Instance

	switch filepath.Ext(filename) {
	case ".cue", ".json":
		cfg, err = r.Compile(filename, nil)
	case ".yaml", ".yml":
		cfg, err = yaml.Decode(r, filename, nil)
	}
	if err != nil {
		return nil, err
	}

	v := inst.Value().Unify(cfg.Value())
	if err := v.Err(); err != nil {
		return nil, err
	}

	c = &Config{}
	if err = v.Decode(c); err != nil {
		return nil, err
	}

	if *verbose && *crd {
		pretty.Print(c)
		fmt.Println()
	}

	return c, nil
}

// fileFromDir computes the openapi json filename from the directory name.
// If filename is not "", it is assumed to be the proto filename in perFile
// mode.
func fileFromDir(dir, filename string) string {
	if filename != "" {
		filename = filename[:len(filename)-len(".proto")]
		return filename + ".json"
	}
	comps := strings.Split(dir, "/")
	if len(comps) == 0 {
		return "istio.json"
	}

	comps = append([]string{"istio"}, comps...)

	return strings.Join(append(comps, "json"), ".")
}

func (c *Config) completeBuildPlan() error {
	root := c.cwd

	buildPlan := c.Directories

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
		case buildPlan[dir][0].Mode == perFile:
			if len(buildPlan[dir][0].ProtoFiles) > 0 {
				buildPlan[dir] = append(buildPlan[dir], Grouping{
					ProtoFiles: []string{file},
					Mode:       perFile,
				})
				break
			}
			fallthrough
		case buildPlan[dir][0].Mode == allFiles:
			buildPlan[dir][0].ProtoFiles = append(buildPlan[dir][0].ProtoFiles, file)
		}

		return nil
	})
	if err != nil {
		return err
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

	return nil
}

func (g *Grouping) update(root, dir string) {
	g.dir = dir

	if g.OapiFilename == "" {
		filename := ""
		if g.Mode == perFile && len(g.ProtoFiles) > 0 {
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
