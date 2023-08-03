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

// genoapi generates OpenAPI files from .proto definitions and other sources
// of CUE constraints. It requires the definition of a configuration file
// that is specified at the Go or CUE module root for which one wishes to
// generate the OpenAPI files.
//
// Generation adopts the Proto <-> JSON mappings conventions. Most notably,
// field names are converted to JSON names.
//
// Generation involves the following steps:
//
//  1. Convert .proto files to CUE files
//  2. Validate the consistency of the CUE defintions
//  3. Convert CUE files to self-contained OpenAPI files.
//
// Each of which is documented in more detail below.
// 1. Converting Proto to CUE
//
// genoapi generates all .proto files using a single builder. As the Istio
// OpenAPI files are self-contained, this is not strictly necessary, but it
// allows for better checking and works better if one wants to generate the
// intermediate CUE results for evaluation.
//
// Field names are mapped using JSON naming.
//
// Protobuf definitions may contain (cue.val) and (cue.opt) options to annotate
// fields with constraints.
//
// Caveats:
//   - It is assumed that the input .proto files are valid and compile with
//     protoc. The conversion may ignore errors if files are invalid.
//   - CUE package names share the same naming conventions as Go packages. CUE
//     requires the go_package option to exist and be well-defined. Note that
//     some of the gogoproto go_package definition are illformed. Be sure to use
//     the original .proto files for the google protobuf types.
//
// 2. Combine and validate generated CUE
//
// CUE files that reside in the same directory as a .proto file and that have
// the same package name as the corresponding Go package are automatically
// merged into the generated CUE definitions. Merging happens based on the
// generated CUE names.
//
// The combines CUE definitions are validated for consistency before proceeding
// to the next step.
// 3. Converting CUE to OpenAPI
//
// In this step a self-contained OpenAPI definition is generated for each
// directory containing proto definitions. Files are made self-contained by
// including a schema definition for each imported type within the OpenAPI
// spec itself. To avoid name collissions, types are, by convention, prefixed
// with their proto package name.
//
// # Possible extensions to the generation pipeline
//
// The generation pipeline can be augmented by injecting CUE from other sources
// before step 2. As combining CUE sources is a commutative operation, order
// of injection does not matter and there is no need for the user to be
// explicit about any order or injection points.
//
// Examples of other possible CUE sources are:
// - constraints extracted from Go code
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/cuecontext"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/openapi"
	"cuelang.org/go/encoding/protobuf"
	"golang.org/x/exp/slices"
	"sigs.k8s.io/yaml"
)

var (
	configFile = flag.String("f", "", "configuration file; by default the directory  in which this file is located is assumed to be the root")
	help       = flag.Bool("help", false, "show documentation for this tool")

	paths   = flag.String("paths", "/protobuf", "comma-separated path to search for .proto imports")
	include = flag.String("include", "", "comma-separated prefixes for files and folders to include when searching for .proto files to process")
	exclude = flag.String("exclude", "", "comma-separated prefixes for files and folders to exclude when searching for .proto files to process")
	verbose = flag.Bool("verbose", false, "print debugging output")

	// Unused, for backwards compat
	_ = flag.Bool("crd", false, "generate CRD validation yaml based on the Istio protos and cue files")

	snake = flag.String("snake", "", "comma-separated fields to add a snake case")

	status = flag.String("status", "", "status field schema name to use for CRDs; only accepted when crd flag is true")

	frontMatterMap map[string][]string
)

const (
	usage = `cue-gen generates OpenAPI definitions and CRDs from Protobuf definitions

Usage:

	cue-gen -paths=<proto-include>,... -f <config> <flags>

Flags:
`

	helpTxt = `
cue-gen converts Protobuf definitions to OpenAPI using hermetic CUE definitions
as an intermediate representation. The .proto files may be annotated to add
additional constraints (see cuelang.org/go/encoding/protobuf).

IMPORTANT:
The -path command line flags must include directories for imports in proto files
that are not located within the Go module. The path for Google protobuf files
imports should thereby precede the path for gogo-proto files. The latter has
invalid copies of the former that will break the build if they are selected
over to original Google files.


Configuration File

The configuration file has the following format, expressed in CUE:

%s
`
)

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	if *help {
		b := cueDoc
		if split := bytes.Split(b, []byte("\n\n")); len(split) > 2 {
			b = bytes.Join(split[2:], []byte("\n\n"))
		}
		fmt.Print(usage)
		flag.PrintDefaults()

		fmt.Printf(helpTxt, b)
		return
	}

	cwd, _ := os.Getwd()

	// Include paths for proto evaluation. Take relative paths relative to the
	// current working directory.
	importPaths := []string{cwd} // change cwd later if root is set.
	for _, p := range strings.Split(*paths, ",") {
		if p != "" {
			abs, _ := filepath.Abs(p)
			importPaths = append(importPaths, abs)
		}
	}

	var (
		includes []string
		excludes []string
	)

	if *include != "" {
		includes = strings.Split(*include, ",")
	}
	if *exclude != "" {
		excludes = strings.Split(*exclude, ",")
	}

	if *configFile == "" {
		log.Fatalf("Must specify configuration with the -f option")
	}

	c, err := loadConfig(*configFile)
	if err != nil {
		fatal(err, "Error parsing configuration file")
	}

	dir, err := filepath.Abs(*configFile)
	if err != nil {
		log.Fatalf("Invalid root %q: %v", dir, err)
	}
	cwd, _ = filepath.Split(dir)

	importPaths[0] = cwd

	b := protobuf.NewExtractor(&protobuf.Config{
		Root:   cwd,
		Module: c.Module,
		Paths:  importPaths,
	})

	_ = filepath.Walk(cwd, func(path string, f os.FileInfo, _ error) (err error) {
		if !strings.HasSuffix(path, ".proto") {
			return nil
		}
		// Ignore non-included files
		for _, i := range includes {
			if !strings.HasPrefix(path, cwd+i) {
				return nil
			}
		}
		// Ignore explicitly excluded files
		for _, e := range excludes {
			if strings.HasPrefix(path, cwd+e) {
				return nil
			}
		}
		// skip the imported protos to avoid circular dependency.
		for _, i := range importPaths[1:] {
			if strings.HasPrefix(path, i) {
				return nil
			}
		}
		// Complete the CrdConfig using the CRD annotations.
		c.getCrdConfig(path)
		return b.AddFile(path, nil)
	})

	files, err := b.Files()
	if err != nil {
		fatal(err, "Error generating CUE from proto")
	}

	snakeFields := strings.Split(*snake, ",")

	c.cwd = cwd

	overlay := map[string]load.Source{}
	for _, f := range files {
		filename := f.Filename
		relPath, _ := filepath.Rel(cwd, filename)
		filename = filepath.Join(cwd, relPath)
		// TODO: remove fix snake case here
		// temporary solution to accommodate for fields that can be
		// used in snake cases.
		if len(snakeFields) > 0 {
			fixSnakes(f, snakeFields)
		}
		overlay[filename] = load.FromFile(f)
	}

	// Generate the OpenAPI
	protoNames := map[string]string{}

	// build map of CUE package import paths (Go paths) to proto package paths.
	builds, _ := b.Instances()
	for _, i := range builds {
		protoNames[i.ImportPath] = i.DisplayPath
	}

	builder := &builder{
		Config:     c,
		protoNames: protoNames,
		overlay:    overlay,
	}

	// Build the OpenAPI files.
	if c.Crd == nil {
		log.Fatalf("Must specify the crd section in the configuration")
	}
	if c.Crd.Filename == "" {
		c.Crd.Filename = "customresourcedefinitions"
	}
	if c.Crd.Dir == "" {
		c.Crd.Dir = "kubernetes"
	}
	builder.genCRD()
}

type builder struct {
	*Config
	protoNames map[string]string
	overlay    map[string]load.Source
}

func (x *builder) genCRD() {
	cfg := &load.Config{
		Dir:     x.cwd,
		Module:  x.Module,
		Overlay: x.overlay,
	}

	// Filter down to directories containing definitions for CRDs
	dirset := map[string]struct{}{}
	for _, v := range x.Crd.CrdConfigs {
		for _, d := range v.Directories {
			wd, _ := os.Getwd()
			rp, _ := filepath.Rel(wd, filepath.Dir(d))
			if rp[0] != '.' {
				rp = "./" + rp
			}
			dirset[rp] = struct{}{}
		}
	}
	dirs := []string{}
	for k := range dirset {
		dirs = append(dirs, k)
	}
	sort.Strings(dirs)

	instances := load.Instances(dirs, cfg)

	frontMatterMap = make(map[string][]string)
	extractFrontMatter(instances, frontMatterMap)

	c := cuecontext.New()
	all, err := c.BuildInstances(instances)
	if err != nil {
		fatal(err, "Instance failed")
	}

	schemas := map[string]cue.Value{}
	for _, inst := range all {
		items, err := x.genOpenAPI(inst.BuildInstance().ImportPath, inst)
		if err != nil {
			fatal(err, "Error generating OpenAPI schema")
		}
		schema := items.LookupPath(cue.ParsePath("components.schemas"))
		i, err := schema.Fields()
		if err != nil {
			fatal(err, "cannot iterate")
		}
		for i.Next() {
			key := i.Label()
			schemas[key] = i.Value()
		}
	}

	statusSchema, ok := schemas[*status]
	if !ok && *status != "" {
		fmt.Println("Status schema not found, thus not including status schema in CRDs.")
	}

	for c, v := range x.Config.Crd.CrdConfigs {

		group := c[:strings.LastIndex(c, ".")]
		tp := crdToType[c]

		versionSchemas := map[string]cue.Value{}

		for _, version := range v.CustomResourceDefinition.Spec.Versions {
			var schemaName string
			if n, ok := v.VersionToSchema[version.Name]; ok {
				schemaName = n
			} else {
				schemaName = fmt.Sprintf("%v.%v.%v", group, version.Name, tp)
			}
			sc, ok := schemas[schemaName]
			if !ok {
				log.Fatalf("cannot find schema for %v", schemaName)
			}
			versionSchemas[version.Name] = sc
		}

		completeCRD(v.CustomResourceDefinition, versionSchemas, statusSchema, v.PreserveUnknownFields)
	}

	x.writeCRDFiles()
}

func selectorLabel(sel cue.Selector) string {
	if sel.Type().ConstraintType() == cue.PatternConstraint {
		return "*"
	}
	switch sel.LabelType() {
	case cue.StringLabel:
		return sel.Unquoted()
	case cue.DefinitionLabel:
		return sel.String()[1:]
	}
	// We shouldn't get anything other than non-hidden
	// fields and definitions because we've not asked the
	// Fields iterator for those or created them explicitly.
	panic(fmt.Sprintf("unreachable %v", sel.Type()))
}

func (x *builder) genOpenAPI(name string, inst cue.Value) (cue.Value, error) {
	fmt.Printf("Building OpenAPIs for %s...\n", name)

	if err := inst.Value().Validate(); err != nil {
		fatal(err, "Validation failed.")
	}

	cfg := x.Openapi
	cfg.NameFunc = func(val cue.Value, path cue.Path) string {
		sels := path.Selectors()
		labels := make([]string, len(sels))
		for i, sel := range sels {
			labels[i] = selectorLabel(sel)
		}
		return x.reference(val.BuildInstance().ImportPath, labels)
	}
	cfg.DescriptionFunc = func(v cue.Value) string {
		n := strings.Split(inst.BuildInstance().ImportPath, "/")
		l, _ := v.Label()
		l = l[1:] // Remove leading '#' from definition
		schema := "istio." + n[len(n)-2] + "." + n[len(n)-1] + "." + l
		if res, ok := frontMatterMap[schema]; ok {
			return res[0] + " See more details at: " + res[1]
		}
		// get the first sentence out of the paragraphs.
		for _, doc := range v.Doc() {
			if doc.Text() == "" {
				continue
			}
			if strings.HasPrefix(doc.Text(), "$hide_from_docs") {
				return ""
			}
			if paras := strings.Split(doc.Text(), "\n"); len(paras) > 0 {
				words := strings.Split(paras[0], " ")
				for i, w := range words {
					if strings.HasSuffix(w, ".") {
						return strings.Join(words[:i+1], " ")
					}
				}
			}
		}
		return ""
	}
	// CRD schema does not allow $ref fields.
	cfg.ExpandReferences = true
	file, err := openapi.Generate(inst, cfg)
	if err != nil {
		return cue.Value{}, err
	}
	ctx := cuecontext.New()
	val := ctx.BuildFile(file).Value()
	return val, nil
}

// reference defines the references format based on the protobuf naming.
func (x *builder) reference(goPkg string, path []string) string {
	name := strings.Join(path, ".")

	pkg := x.protoNames[goPkg]
	if pkg == "" {
		// Not a proto package, expand in place.
		return ""
	}
	// Map CUE names to proto names.
	name = strings.Replace(name, "_", ".", -1)
	return pkg + "." + name
}

// extracts the front comments in istio protos.
func extractFrontMatter(ins []*build.Instance, m map[string][]string) {
	const schemaTag = "$schema:"
	const descriptionTag = "$description:"
	const locationTag = "$location:"
	for _, i := range ins {
		for _, f := range i.Files {
			for _, c := range f.Comments() {
				txt := c.Text()
				if strings.HasPrefix(txt, "$") {
					lines := strings.Split(txt, "\n")
					var description, location string
					var schemas []string
					for _, l := range lines {
						l = strings.TrimSpace(l)

						if strings.HasPrefix(l, schemaTag) {
							schemas = append(schemas, strings.TrimSpace(strings.TrimPrefix(l, schemaTag)))
						} else if strings.HasPrefix(l, descriptionTag) {
							description = strings.TrimSpace(strings.TrimPrefix(l, descriptionTag))
						} else if strings.HasPrefix(l, locationTag) {
							location = strings.TrimSpace(strings.TrimPrefix(l, locationTag))
						}
					}
					for _, s := range schemas {
						m[s] = []string{description, location}
					}
				}
			}
		}
	}
}

func fixSnakes(f *ast.File, sf []string) {
	astutil.Apply(f, func(bc astutil.Cursor) bool {
		n := bc.Node()
		switch x := n.(type) {
		case *ast.Field:
			if s := snakeField(x, sf); s != nil {
				bc.InsertAfter(s)
			}
			return true
		default:
			return true
		}
	}, nil)
}

// snakeField returns a Field with snake_case naming, if the field is in
// the list provided.
func snakeField(f *ast.Field, sf []string) *ast.Field {
	if n, i, _ := ast.LabelName(f.Label); !i || !slices.Contains(sf, n) {
		return nil
	}
	snakeCase := protobufName(f)
	if snakeCase == "" {
		return nil
	}
	snaked := &ast.Field{
		Label:    ast.NewIdent(snakeCase),
		Optional: f.Optional,
		Value:    f.Value,
		Attrs:    f.Attrs,
	}
	astutil.CopyMeta(snaked, f) // copy comments and relative positions
	return snaked
}

// protobufName returns the proto name of the given Field.
func protobufName(f *ast.Field) string {
	for _, attr := range f.Attrs {
		if strings.HasPrefix(attr.Text, "@protobuf") {
			for _, str := range strings.Split(attr.Text[10:len(attr.Text)-1], ",") {
				if strings.HasPrefix(str, "name=") {
					return str[5:]
				}
			}
		}
	}
	return ""
}

func (x *builder) writeCRDFiles() {
	dirPath := filepath.Join(x.cwd, x.Config.Crd.Dir)
	// ensure the directory exists
	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		log.Fatalf("Cannot create directory for CRD output: %v", err)
	}
	filename := fmt.Sprintf("%v.gen.yaml", x.Config.Crd.Filename)
	path := filepath.Join(x.cwd, x.Config.Crd.Dir, filename)
	fmt.Printf("Writing CRDs into %v...\n", path)
	out, err := os.Create(path)
	if err != nil {
		log.Fatalf("Cannot create file %v: %v", path, err)
	}
	defer out.Close()

	if _, err = out.WriteString("# DO NOT EDIT - Generated by Cue OpenAPI generator based on Istio APIs.\n"); err != nil {
		log.Fatal(err)
	}

	// sort the configs so that the order is deterministic.
	var keyList []string
	for k := range x.Crd.CrdConfigs {
		keyList = append(keyList, k)
	}
	sort.Strings(keyList)

	for _, k := range keyList {
		y, err := yaml.Marshal(x.Crd.CrdConfigs[k].CustomResourceDefinition)
		if err != nil {
			log.Fatalf("Error marsahling CRD to yaml: %v", err)
		}

		// remove the status and creationTimestamp fields from the output. Ideally we could use OrderedMap to remove those.
		y = bytes.ReplaceAll(y, []byte(statusOutput), []byte(""))
		y = bytes.ReplaceAll(y, []byte(creationTimestampOutput), []byte(""))
		// keep the quotes in the output which is required by helm.
		y = bytes.ReplaceAll(y, []byte("helm.sh/resource-policy: keep"), []byte(`"helm.sh/resource-policy": keep`))
		n, err := out.Write(append(y, []byte("\n---\n")...))
		if err != nil {
			log.Fatalf("Error writing to yaml file: %v", err)
		}
		if n < len(y) {
			log.Fatalf("Error writing to yaml file: %v", io.ErrShortWrite)
		}
	}
}

func fatal(err error, msg string) {
	errors.Print(os.Stderr, err, nil)
	_ = log.Output(2, msg)
	os.Exit(1)
}
