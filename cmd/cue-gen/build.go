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
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	enableCRDGenTag = "+cue-gen"
)

// A mapping from CRD name to proto type name
var crdToType map[string]string

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
	CrdConfigs map[string]*CrdConfig
}

// CrdConfig contains the CRD for each proto type to be generated.
type CrdConfig struct {
	// Contains all directories of source schemas.
	Directories []string
	// Optional. Mapping of version to schema name if schema name
	// not following the <package>.<version>.<name> format.
	VersionToSchema map[string]string

	// Optional. Mapping of version to slice of field paths that
	// need to be marked as `x-kubernetes-preserve-unknown-fields`
	PreserveUnknownFields map[string][]string

	CustomResourceDefinition *apiext.CustomResourceDefinition
}

func loadConfig(filename string) (c *Config, err error) {
	r := &cue.Runtime{} //nolint:staticcheck

	f, err := docCueBytes()
	if err != nil {
		return nil, err
	}
	inst, err := r.Compile("doc.cue", f) //nolint:staticcheck
	if err != nil {
		log.Fatal(err)
	}

	var cfg *cue.Instance

	switch filepath.Ext(filename) {
	case ".cue", ".json":
		cfg, err = r.Compile(filename, nil) //nolint:staticcheck
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

	if c.Crd == nil {
		c.Crd = &CrdGen{}
	}
	if c.Crd.CrdConfigs == nil {
		c.Crd.CrdConfigs = map[string]*CrdConfig{}
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
		return filename + ".gen.json"
	}
	comps := strings.Split(dir, "/")
	if len(comps) == 0 {
		return "istio.gen.json"
	}

	comps = append([]string{"istio"}, comps...)

	return strings.Join(append(comps, "gen.json"), ".")
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

func (c *Config) getCrdConfig(filename string) {
	prefix := findTypePrefix(filename)
	for _, p := range protoElems(filename) {
		switch x := p.(type) {
		case *proto.Message:
			if x.Comment == nil {
				continue
			}
			out := extractCrdTags(x.Comment.Lines, prefix)
			for t, v := range out {
				if _, ok := c.Crd.CrdConfigs[t]; !ok {
					c.Crd.CrdConfigs[t] = &CrdConfig{
						VersionToSchema:          map[string]string{},
						CustomResourceDefinition: &apiext.CustomResourceDefinition{},
						PreserveUnknownFields:    map[string][]string{},
					}
				}
				c.Crd.CrdConfigs[t].Directories = append(c.Crd.CrdConfigs[t].Directories, filename)
				d := c.Crd.CrdConfigs[t]
				convertCrdConfig(v, t, d)
				if crdToType == nil {
					crdToType = map[string]string{}
				}
				crdToType[t] = x.Name
			}
		}
	}
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

// find the prefix of the type in FQDN, e.g. istio.networking
func findTypePrefix(filename string) string {
	for _, d := range protoElems(filename) {
		switch x := d.(type) {
		case *proto.Package:
			if strings.LastIndex(x.Name, ".") == -1 {
				return x.Name
			}
			// as an Istio api naming convention, we strip out the last element.
			return x.Name[:strings.LastIndex(x.Name, ".")]
		}
	}
	return ""
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

func extractCrdTags(lines []string, prefix string) map[string]map[string]string {
	lines = cleanComments(lines)
	out := map[string]map[string]string{}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		s := strings.SplitN(line[len(enableCRDGenTag+":"):], ":", 2)
		if len(s) < 2 {
			log.Fatalf("cannot recognize type from line: %v", line)
			continue
		}

		t := prefix + "." + s[0]
		kv := strings.SplitN(s[1], ":", 2)
		var k, v string
		if len(kv) == 2 {
			k = kv[0]
			v = kv[1]
		} else if len(kv) == 1 {
			k = kv[0]
			v = ""
		} else {
			log.Fatalf("cannot retrieve config key value pair from line: %v", line)
			continue
		}
		if _, ok := out[t]; !ok {
			c := map[string]string{}
			out[t] = c
		}
		if k == "printerColumn" {
			out[t][k] += ";;" + v
		} else {
			out[t][k] = v
		}
	}
	return out
}

func convertCrdConfig(c map[string]string, t string, cfg *CrdConfig) {
	t = t[strings.LastIndex(t, ".")+1:]
	src := cfg.CustomResourceDefinition
	src.Spec.Names = apiext.CustomResourceDefinitionNames{
		Kind:     t,
		ListKind: t + "List",
		Singular: strings.ToLower(t),
		Plural:   strings.ToLower(t) + "s",
	}
	version := apiext.CustomResourceDefinitionVersion{
		Served: true,
	}
	var sc string
	for k, v := range c {
		switch k {
		case "groupName":
			src.Spec.Group = v
		case "scope":
			if v == "Namespaced" {
				src.Spec.Scope = apiext.NamespaceScoped
			} else {
				src.Spec.Scope = apiext.ClusterScoped
			}
		case "resource":
			mp := extractKeyValue(v)
			for n, m := range mp {
				switch n {
				case "categories":
					src.Spec.Names.Categories = appendSlice(src.Spec.Names.Categories, strings.Split(m, ","))
				case "plural":
					src.Spec.Names.Plural = m
				case "kind":
					src.Spec.Names.Kind = m
				case "shortNames":
					src.Spec.Names.ShortNames = appendSlice(src.Spec.Names.ShortNames, strings.Split(m, ","))
				case "singular":
					src.Spec.Names.Singular = m
				case "listKind":
					src.Spec.Names.ListKind = m
				}
			}
		case "annotations":
			src.Annotations = appendMap(src.Annotations, extractKeyValue(v))
		case "labels":
			src.Labels = appendMap(src.Labels, extractKeyValue(v))
		case "subresource":
			if v == "status" {
				version.Subresources = &apiext.CustomResourceSubresources{Status: &apiext.CustomResourceSubresourceStatus{}}
			}
		case "storageVersion":
			version.Storage = true
		case "printerColumn":
			pcs := strings.Split(v, ";;")
			for _, pc := range pcs {
				if pc == "" {
					continue
				}
				column := apiext.CustomResourceColumnDefinition{}
				for n, m := range extractKeyValue(pc) {
					switch n {
					case "name":
						column.Name = m
					case "type":
						column.Type = m
					case "description":
						column.Description = m
					case "JSONPath":
						column.JSONPath = m
					}
				}
				version.AdditionalPrinterColumns = append(version.AdditionalPrinterColumns, column)
			}
		case "version":
			version.Name = v
		case "schema":
			sc = v
		}
	}
	if sc != "" {
		m := cfg.VersionToSchema
		m[version.Name] = sc
		cfg.VersionToSchema = m
	}

	// store the fields to mark as preserved in the config
	if f, ok := c["preserveUnknownFields"]; ok {
		cfg.PreserveUnknownFields[version.Name] = strings.Split(f, ",")
	}

	src.Spec.Versions = append(src.Spec.Versions, version)
	src.Name = fmt.Sprintf("%v.%v", src.Spec.Names.Plural, src.Spec.Group)
}

// extractkeyValue extracts a string to key value pairs
// e.g. a=b,b=c to map[a:b b:c]
// and a=b,c,d,e=f to map[a:b,c,d e:f]
func extractKeyValue(s string) map[string]string {
	out := map[string]string{}
	if s == "" {
		return out
	}
	splits := strings.Split(s, "=")
	if len(splits) == 1 {
		out[splits[0]] = ""
	}
	if strings.Contains(splits[0], ",") {
		log.Fatalf("cannot parse %v to key value pairs", s)
	}
	nextkey := splits[0]
	for i := 1; i < len(splits); i++ {
		if splits[i] == "" || splits[i] == "," {
			log.Fatalf("cannot parse %v to key value paris, invalid value", s)
		}
		if !strings.Contains(splits[i], ",") && i != len(splits)-1 {
			log.Fatalf("cannot parse %v to key value pairs, missing separator", s)
		}
		if i == len(splits)-1 {
			out[nextkey] = strings.Trim(splits[i], "\"'`")
			continue
		}
		index := strings.LastIndex(splits[i], ",")
		out[nextkey] = strings.Trim(splits[i][:index], "\"'`")
		nextkey = splits[i][index+1:]
		if nextkey == "" {
			log.Fatalf("cannot parse %v to key value pairs, missing key", s)
		}
	}
	return out
}

func cleanComments(lines []string) []string {
	out := []string{}
	var prevLine string
	for _, line := range lines {
		line = strings.Trim(line, " ")

		if line == "-->" {
			out = append(out, prevLine)
			prevLine = ""
			continue
		}

		if !strings.HasPrefix(line, enableCRDGenTag) {
			if prevLine != "" && len(line) != 0 {
				prevLine += " " + line
			}
			continue
		}

		out = append(out, prevLine)

		prevLine = line

	}
	if prevLine != "" {
		out = append(out, prevLine)
	}
	return out
}

func appendSlice(dst []string, src []string) []string {
	for _, e := range src {
		if !func(els []string, el string) bool {
			for _, o := range els {
				if o == el {
					return true
				}
			}
			return false
		}(dst, e) {
			dst = append(dst, e)
		}
	}
	return dst
}

func appendMap(dst map[string]string, src map[string]string) map[string]string {
	if dst == nil {
		dst = map[string]string{}
	}
	for k, v := range src {
		dst[k] = v
	}
	return dst
}
