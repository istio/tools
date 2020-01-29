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
//   1. Convert .proto files to CUE files
//   2. Validate the consistency of the CUE defintions
//   3. Convert CUE files to self-contained OpenAPI files.
//
// Each of which is documented in more detail below.
//
//
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
// - It is assumed that the input .proto files are valid and compile with
//   protoc. The conversion may ignore errors if files are invalid.
// - CUE package names share the same naming conventions as Go packages. CUE
//   requires the go_package option to exist and be well-defined. Note that
//   some of the gogoproto go_package definition are illformed. Be sure to use
//   the original .proto files for the google protobuf types.
//
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
//
//
// 3. Converting CUE to OpenAPI
//
// In this step a self-contained OpenAPI definition is generated for each
// directory containing proto definitions. Files are made self-contained by
// including a schema definition for each imported type within the OpenAPI
// spec itself. To avoid name collissions, types are, by convention, prefixed
// with their proto package name.
//
//
// Possible extensions to the generation pipeline
//
// The generation pipeline can be augmented by injecting CUE from other sources
// before step 2. As combining CUE sources is a commutative operation, order
// of injection does not matter and there is no need for the user to be
// explicit about any order or injection points.
//
// Examples of other possible CUE sources are:
// - constraints extracted from Go code
//
package main

//go:generate go-bindata --nocompress --nometadata --pkg main -o assets.gen.go doc.cue

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/ast"
	"cuelang.org/go/cue/ast/astutil"
	"cuelang.org/go/cue/build"
	"cuelang.org/go/cue/errors"
	"cuelang.org/go/cue/format"
	"cuelang.org/go/cue/load"
	"cuelang.org/go/encoding/openapi"
	"cuelang.org/go/encoding/protobuf"
	"github.com/emicklei/proto"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/ghodss/yaml"
)

var (
	configFile = flag.String("f", "", "configuration file; by default the directory  in which this file is located is assumed to be the root")
	help       = flag.Bool("help", false, "show documentation for this tool")

	inplace = flag.Bool("inplace", false, "generate configurations in place")
	paths   = flag.String("paths", "/protobuf", "comma-separated path to search for .proto imports")
	verbose = flag.Bool("verbose", false, "print debugging output")

	// manually configuring builds
	all = flag.Bool("all", false, "combine all the matched outputs in a single file; the 'all' section must be specified in the configuration")

	crd = flag.Bool("crd", false, "generate CRD validation yaml based on the Istio protos and cue files")

	snake = flag.String("snake", "", "comma-separated fields to add a snake case")

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

The configuration file has the followign format, expressed in CUE:

%s
`
)

func main() {
	flag.Parse()
	log.SetFlags(log.Lshortfile)

	if *help {
		b, err := docCueBytes()
		if err != nil {
			log.Fatal(err)
		}
		if split := bytes.Split(b, []byte("\n\n")); len(split) > 2 {
			b = bytes.Join(split[2:], []byte("\n\n"))
		}
		fmt.Println(usage)
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
		// skip the imported protos to avoid circular dependency.
		for _, i := range importPaths[1:] {
			if strings.HasPrefix(path, i) {
				return nil
			}
		}
		// Complete the CrdConfig using the CRD annotations.
		if *crd {
			c.getCrdConfig(path)
		}
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
		if *crd && len(snakeFields) > 0 {
			fixSnakes(f, snakeFields)
		}
		overlay[filename] = load.FromFile(f)

		if *inplace {
			b, err := format.Node(f)
			if err != nil {
				fatal(err, "Error formatting file: ")
			}
			_ = os.MkdirAll(filepath.Dir(filename), 0755)
			if err := ioutil.WriteFile(filename, b, 0644); err != nil {
				log.Fatalf("Error writing file: %v", err)
			}
		}

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
	if *all {
		if c.All == nil {
			log.Fatalf("Must specify the all section in the configuration")
		}
		builder.genAll(c.All)
	} else if *crd {
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
	} else {
		err := c.completeBuildPlan()
		if err != nil {
			log.Fatalf("Error completing build plan: %v", err)
		}
		for dir, groupings := range c.Directories {
			for _, g := range groupings {
				builder.gen(dir, &g)
			}
		}
	}
}

type builder struct {
	*Config
	protoNames map[string]string
	overlay    map[string]load.Source
}

func (x *builder) gen(dir string, g *Grouping) {
	cfg := &load.Config{
		Dir:     x.cwd,
		Module:  x.Module,
		Overlay: x.overlay,
	}

	instances := load.Instances([]string{"./" + dir}, cfg)
	inst := cue.Build(instances)[0]
	if inst.Err != nil {
		fatal(inst.Err, "Instance failed")
	}

	schemas, err := x.genOpenAPI(g.OapiFilename, inst)
	if err != nil {
		fatal(err, "Error generating OpenAPI file")
	}

	if g.Mode != allFiles {
		x.filterOpenAPI(schemas, g)
	}

	x.writeOpenAPI(schemas, g)
}

func (x *builder) genAll(g *Grouping) {
	cfg := &load.Config{
		Dir:     x.cwd,
		Module:  x.Module,
		Overlay: x.overlay,
	}

	instances := load.Instances([]string{"./..."}, cfg)
	all := cue.Build(instances)
	for _, inst := range all {
		if inst.Err != nil {
			fatal(inst.Err, "Instance failed")
		}
	}

	found := map[string]bool{}

	schemas := &openapi.OrderedMap{}

	for _, inst := range all {
		items, err := x.genOpenAPI(inst.ImportPath, inst)
		if err != nil {
			fatal(err, "Error generating OpenAPI file")
		}
		for _, kv := range items.Pairs() {
			if found[kv.Key] {
				continue
			}
			found[kv.Key] = true

			schemas.Set(kv.Key, kv.Value)
		}
	}

	x.writeOpenAPI(schemas, g)
}

func (x *builder) genCRD() {
	cfg := &load.Config{
		Dir:     x.cwd,
		Module:  x.Module,
		Overlay: x.overlay,
	}

	instances := load.Instances([]string{"./..."}, cfg)

	frontMatterMap = make(map[string][]string)
	extractFrontMatter(instances, frontMatterMap)

	all := cue.Build(instances)
	for _, inst := range all {
		if inst.Err != nil {
			fatal(inst.Err, "Instance failed")
		}
	}

	schemas := map[string]*openapi.OrderedMap{}
	for _, inst := range all {
		items, err := x.genOpenAPI(inst.ImportPath, inst)
		if err != nil {
			fatal(err, "Error generating OpenAPI schema")
		}
		for _, kv := range items.Pairs() {
			schemas[kv.Key] = kv.Value.(*openapi.OrderedMap)
		}
	}

	for c, v := range x.Config.Crd.CrdConfigs {

		group := c[:strings.LastIndex(c, ".")]
		tp := crdToType[c]

		versionSchemas := map[string]*openapi.OrderedMap{}

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

		completeCRD(v.CustomResourceDefinition, versionSchemas)
	}

	x.writeCRDFiles()
}

func (x *builder) genOpenAPI(name string, inst *cue.Instance) (*openapi.OrderedMap, error) {
	fmt.Printf("Building OpenAPIs for %s...\n", name)

	if err := inst.Value().Validate(); err != nil {
		fatal(err, "Validation failed.")
	}

	gen := *x.Openapi
	gen.ReferenceFunc = func(p *cue.Instance, path []string) string {
		return x.reference(p.ImportPath, path)
	}

	gen.DescriptionFunc = func(v cue.Value) string {
		if *crd {
			n := strings.Split(inst.ImportPath, "/")
			l, _ := v.Label()
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
		for _, doc := range v.Doc() {
			if doc.Text() == "" {
				continue
			}
			// Cut off first section, but don't stop if this ends with
			// an example, list, or the like, as it will end weirdly.
			// Also remove any protoc-gen-docs annotations at the beginning
			// and any new-line.
			split := strings.Split(doc.Text(), "\n\n")
			k := 1
			for ; k < len(split) && strings.HasSuffix(split[k-1], ":"); k++ {
			}
			s := strings.Fields(strings.Join(split[:k], "\n"))
			for i := 0; i < len(s) && strings.HasPrefix(s[i], "$"); i++ {
				s[i] = ""
			}
			return strings.Join(s, " ")
		}
		return ""
	}

	if *crd {
		// CRD schema does not allow $ref fields.
		gen.ExpandReferences = true
	}

	return gen.Schemas(inst)
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

// filterOpenAPI filters out unneeded elements from a generated OpenAPI.
// It does so my looking up the top-level items in the proto files defined
// in g, recursively marking their dependencies, and then eliminating any
// schema from items that was not marked.
func (x *builder) filterOpenAPI(items *openapi.OrderedMap, g *Grouping) {
	// All references found.
	m := marker{
		found:   map[string]bool{},
		schemas: items,
	}

	// Get top-level definitions for the files in the given Grouping
	for _, f := range g.ProtoFiles {
		goPkg := ""
		for _, e := range protoElems(filepath.Join(x.cwd, g.dir, f)) {
			switch v := e.(type) {
			case *proto.Option:
				if v.Name == "go_package" {
					goPkg, _ = strconv.Unquote(v.Constant.SourceRepresentation())
				}
			case *proto.Message:
				m.markReference(x.reference(goPkg, []string{v.Name}))

			case *proto.Enum:
				m.markReference(x.reference(goPkg, []string{v.Name}))
			}
		}
	}

	// Now eliminate unused top-level items.
	k := 0
	pairs := m.schemas.Pairs()
	for i := 0; i < len(pairs); i++ {
		if m.found[pairs[i].Key] {
			pairs[k] = pairs[i]
			k++
		}
	}
	m.schemas.SetAll(pairs[:k])
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
	if n, i, _ := ast.LabelName(f.Label); !i || !contains(sf, n) {
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

func (x *builder) writeOpenAPI(schemas *openapi.OrderedMap, g *Grouping) {
	oapi := &openapi.OrderedMap{}
	oapi.Set("openapi", "3.0.0")

	info := &openapi.OrderedMap{}
	info.Set("title", g.Title)
	info.Set("version", g.Version)
	oapi.Set("info", info)

	comps := &openapi.OrderedMap{}
	comps.Set("schemas", schemas)
	oapi.Set("components", comps)

	b, _ := json.Marshal(oapi)

	// Note: this just tests basic OpenAPI 3 validity. It cannot, of course,
	// know if the the proto files were correctly mapped.
	_, err := openapi3.NewSwaggerLoader().LoadSwaggerFromData(b)
	if err != nil {
		log.Fatalf("Invalid OpenAPI generated: %v", err)
	}

	var buf bytes.Buffer
	_ = json.Indent(&buf, b, "", "  ")

	filename := filepath.Join(x.cwd, g.dir, g.OapiFilename)
	fmt.Printf("Writing OpenAPI schemas into %v...\n", filename)
	err = ioutil.WriteFile(filename, buf.Bytes(), 0644)
	if err != nil {
		log.Fatalf("Error writing OpenAPI file %s in dir %s: %v", g.OapiFilename, g.dir, err)
	}
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
		// TODO(jasonwzm): remove this when we no longer support Kubernetes <=1.13. Nullable is a valid field in structural
		// schema, but we are removing it to accommodate for JSONSchema in Kubernetes <=1.13.
		re := regexp.MustCompile(`\s*nullable:\strue`)
		y = re.ReplaceAll(y, []byte(""))

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

type marker struct {
	found   map[string]bool
	schemas *openapi.OrderedMap
}

func (x *marker) markReference(ref string) {
	if x.found[ref] {
		return
	}
	x.found[ref] = true

	for _, kv := range x.schemas.Pairs() {
		if kv.Key == ref {
			x.markRecursive(kv.Value.(*openapi.OrderedMap))
			return
		}
	}
	panic("should not happen")
}

func (x *marker) markRecursive(m *openapi.OrderedMap) {
	for _, kv := range m.Pairs() {
		switch v := kv.Value.(type) {
		case *openapi.OrderedMap:
			x.markRecursive(v)
		case []*openapi.OrderedMap:
			for _, m := range v {
				x.markRecursive(m)
			}
		case string:
			if kv.Key == "$ref" {
				x.markReference(kv.Value.(string)[len("#/components/schemas/"):])
			}
		}
	}
}

func fatal(err error, msg string) {
	errors.Print(os.Stderr, err, nil)
	_ = log.Output(2, msg)
	os.Exit(1)
}

func contains(l []string, s string) bool {
	for _, e := range l {
		if e == s {
			return true
		}
	}
	return false
}
