// Copyright Istio Authors
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

// A simple program that consumes a YAML file describing Kubernetes resource annotations and produces as output
// a Go source file providing references to those annotations, and an HTML documentation file describing those
// annotations (for use on istio.io)

package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

const (
	outputTemplate = `
// GENERATED FILE -- DO NOT EDIT

package {{ .Package }}

type FeatureStatus int

const (
	Alpha FeatureStatus = iota
	Beta
	Stable
)

func (s FeatureStatus) String() string {
	switch s {
	case Alpha:
		return "Alpha"
	case Beta:
		return "Beta"
	case Stable:
		return "Stable"
	}
	return "Unknown"
}

type ResourceTypes int

const (
	Unknown ResourceTypes = iota
	{{- range .KnownTypes }}
    {{ . }}
    {{- end }}
)

func (r ResourceTypes) String() string {
	switch r {
	{{- range $i, $t := .KnownTypes }}
	case {{ add $i 1 }}:
		return "{{$t}}"
	{{- end }}
	}
	return "Unknown"
}

// Instance describes a single resource {{ .Collection.NameLowercase }}
type Instance struct {
	// The name of the {{ .Collection.NameLowercase }}.
	Name string

	// Description of the {{ .Collection.NameLowercase }}.
	Description string

	// FeatureStatus of this {{ .Collection.NameLowercase }}.
	FeatureStatus FeatureStatus

	// Hide the existence of this {{ .Collection.NameLowercase }} when outputting usage information.
	Hidden bool

	// Mark this {{ .Collection.NameLowercase }} as deprecated when generating usage information.
	Deprecated bool

	// The types of resources this {{ .Collection.NameLowercase }} applies to.
	Resources []ResourceTypes
}

var (
{{ range .Variables }}
	{{ .GoName }} = Instance {
		Name:          "{{ .Name }}",
		Description:   {{ processGoDescription .Description 24 }},
		FeatureStatus: {{ .FeatureStatus }},
		Hidden:        {{ .Hidden }},
		Deprecated:    {{ .Deprecated }},
		Resources: []ResourceTypes{
			{{- range .Resources }}
			{{ . }},
			{{- end }}
		},
	}
{{ end }}
)

func AllResource{{ .Collection.NamePlural }}() []*Instance {
	return []*Instance {
		{{- range .Variables }}
		&{{ .GoName }},
		{{- end }}
	}
}

func AllResourceTypes() []string {
	return []string {
		{{- range .KnownTypes }}
		"{{ . }}",
		{{- end }}
	}
}
`

	htmlOutputTemplate = `---
title: Resource {{ .Collection.NamePlural }}
description: Resource {{ .Collection.NameLowercasePlural }} used by Istio.
location: {{ .Collection.Link }}
weight: 60
---

<p>
This page presents the various resource <a href="{{ .Collection.ConceptLink }}">{{ .Collection.NameLowercasePlural }}</a> that
Istio supports to control its behavior.
</p>
{{- range .Variables -}}
{{- if not .Hidden }}
<h2 id="{{ .GoName }}">{{ .Name }}</h2>
<table class="annotations">
  <tbody>
    <tr{{ if .Deprecated }} class="deprecated"{{ end }}>
      <th>Name</th>
      <td><code>{{ .Name }}</code></td>
    </tr>
    <tr>
      <th>Feature Status</th>
    {{- if .Deprecated }}
      <td>Deprecated</td>
    {{- else }}
      <td>{{ .FeatureStatus }}</td>
    {{- end }}
    </tr>
    <tr>
      <th>Resource Types</th>
      <td>{{ .Resources }}</td>
    </tr>
    <tr>
      <th>Description</th>
      <td>{{ processHTMLDescription .Description }}</td>
    </tr>
  </tbody>
</table>
{{- end -}}
{{- end -}}
`
)

type FeatureStatus string

const (
	Alpha  FeatureStatus = "Alpha"
	Beta   FeatureStatus = "Beta"
	Stable FeatureStatus = "Stable"
)

// Collection represents template fields for either annotations or labels.
type Collection struct {
	Name                string
	NamePlural          string
	NameLowercase       string
	NameLowercasePlural string
	// Link is the location of the generated page on istio.io.
	Link string
	// ConceptLink is the link to the concept page for the collection type.
	ConceptLink string
}

var (
	annotations = Collection{
		Name:                "Annotation",
		NamePlural:          "Annotations",
		NameLowercase:       "annotation",
		NameLowercasePlural: "annotations",
		Link:                "https://istio.io/docs/reference/config/annotations/",
		ConceptLink:         "https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/",
	}

	labels = Collection{
		Name:                "Label",
		NamePlural:          "Labels",
		NameLowercase:       "label",
		NameLowercasePlural: "labels",
		Link:                "https://istio.io/docs/reference/config/labels/",
		ConceptLink:         "https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/",
	}
)

func collectionForType(typ string) (Collection, error) {
	switch typ {
	case annotations.NameLowercase:
		return annotations, nil
	case labels.NameLowercase:
		return labels, nil
	default:
		return Collection{}, fmt.Errorf("unrecognized variable_type: %s", typ)
	}
}

var (
	input          string
	output         string
	htmlOutput     string
	collectionType string
	collection     Collection

	nameSeparator = regexp.MustCompile(`[._\-]`)

	rootCmd = cobra.Command{
		Use:   "annotations_prep",
		Short: "Generates a Go source file and HTML file containing annotations/labels.",
		Long: "Generates a Go source file and HTML file containing annotation/label definitions based " +
			"on an input YAML file.",
		Run: func(cmd *cobra.Command, args []string) {
			processFlags()
			yamlContent, err := os.ReadFile(input)
			if err != nil {
				log.Fatalf("unable to read input YAML file: %v", err)
			}

			// Unmarshal the file.
			var variables []Variable
			switch collectionType {
			case annotations.NameLowercase:
				var cfg AnnotationConfiguration
				if err = yaml.Unmarshal(yamlContent, &cfg); err != nil {
					log.Fatalf("error parsing input YAML file: %v", err)
				}
				variables = cfg.Variables
			case labels.NameLowercase:
				var cfg LabelConfiguration
				if err = yaml.Unmarshal(yamlContent, &cfg); err != nil {
					log.Fatalf("error parsing input YAML file: %v", err)
				}
				variables = cfg.Variables
			default:
				log.Fatalf("invalid value for collection_type: %s", collectionType)
			}

			// Find all the known resource types
			m := make(map[string]bool)
			for _, a := range variables {
				for _, r := range a.Resources {
					m[r] = true
				}
			}
			knownTypes := make([]string, 0, len(m))
			for k := range m {
				knownTypes = append(knownTypes, k)
			}
			sort.Strings(knownTypes)

			// Process/cleanup the values read in from YAML.
			for i := range variables {
				if variables[i].Name == "" {
					log.Fatalf("variable %d in input YAML file missing name", i)
				}

				// Generate sensible defaults for values if not provided in the yaml.
				variables[i].GoName = generateVariableName(variables[i])
				variables[i].FeatureStatus = generateFeatureStatus(variables[i])
			}

			// sort by name
			sort.Slice(variables, func(i, j int) bool {
				return strings.Compare(variables[i].Name, variables[j].Name) < 0
			})

			// Create the output file template.
			t, err := template.New("varTemplate").Funcs(template.FuncMap{
				"processGoDescription": processGoDescription, "add": add,
			}).Parse(outputTemplate)
			if err != nil {
				log.Fatalf("failed parsing variable template: %v", err)
			}

			// Generate the Go source.
			var goSource bytes.Buffer
			if err := t.Execute(&goSource, map[string]interface{}{
				"Package":    getPackage(),
				"KnownTypes": knownTypes,
				"Variables":  variables,
				"Collection": collection,
			}); err != nil {
				log.Fatalf("failed generating output Go source code %s: %v", output, err)
			}

			if err := os.WriteFile(output, goSource.Bytes(), 0o666); err != nil {
				log.Fatalf("Failed writing to output file %s: %v", output, err)
			}

			if htmlOutput != "" {
				// Create the HTML output file template.
				t, err = template.New("htmlOutputTemplate").Funcs(template.FuncMap{
					"processHTMLDescription": processHTMLDescription,
				}).Parse(htmlOutputTemplate)
				if err != nil {
					log.Fatalf("failed parsing HTML output template: %v", err)
				}

				// Generate the HTML file.
				var htmlFile bytes.Buffer
				if err := t.Execute(&htmlFile, map[string]interface{}{
					"Package":    getPackage(),
					"Variables":  variables,
					"Collection": collection,
				}); err != nil {
					log.Fatalf("failed generating output HTML file %s: %v", output, err)
				}

				if err := os.WriteFile(htmlOutput, htmlFile.Bytes(), 0o666); err != nil {
					log.Fatalf("Failed writing to output file %s: %v", htmlOutput, err)
				}
			}
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&input, "input", "",
		"Input YAML file to be parsed.")
	rootCmd.PersistentFlags().StringVar(&output, "output", "",
		"Output Go file to be generated.")
	rootCmd.PersistentFlags().StringVar(&htmlOutput, "html_output", "",
		"Output HTML file to be generated.")
	rootCmd.PersistentFlags().StringVar(&collectionType, "collection_type", annotations.NameLowercase,
		fmt.Sprintf("Output type for the generated collection. Allowed values are '%s' or '%s'.",
			annotations.NameLowercase, labels.NameLowercase))

	flag.CommandLine.VisitAll(func(gf *flag.Flag) {
		rootCmd.PersistentFlags().AddGoFlag(gf)
	})
}

func processFlags() {
	var err error
	collection, err = collectionForType(collectionType)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

type Variable struct {
	// The name of the generated golang variable.
	GoName string `json:"variableName"`

	// The name of the collection variable.
	Name string `json:"name"`

	// FeatureStatus of the collection variable.
	FeatureStatus string `json:"featureStatus"`

	// Description of the collection variable.
	Description string `json:"description"`

	// Hide the existence of this collection variable when outputting usage information.
	Hidden bool `json:"hidden"`

	// Mark this annotation as deprecated when generating usage information.
	Deprecated bool `json:"deprecated"`

	// Indicates the types of resources this collection variable can be applied to.
	Resources []string `json:"resources"`
}

type AnnotationConfiguration struct {
	Variables []Variable `json:"annotations"`
}

type LabelConfiguration struct {
	Variables []Variable `json:"labels"`
}

func getPackage() string {
	path, _ := filepath.Abs(output)
	return filepath.Base(filepath.Dir(path))
}

func generateVariableName(v Variable) string {
	if len(v.GoName) > 0 {
		return v.GoName
	}

	// Split the annotation name to separate the namespace/name portions.
	parts := strings.Split(v.Name, "/")
	ns := parts[0]
	name := parts[1]

	// First, process the namespace portion ...

	// Strip .istio.io from the namespace portion of the annotation name.
	ns = strings.TrimSuffix(ns, ".istio.io")

	// Separate the words by spaces and capitalize each word.
	ns = strings.ReplaceAll(ns, ".", " ")
	// nolint: staticcheck
	ns = strings.Title(ns)

	// Reverse the namespace words so that they increase in specificity from left to right.
	nsParts := strings.Split(ns, " ")
	ns = ""
	for i := len(nsParts) - 1; i >= 0; i-- {
		ns += nsParts[i]
	}

	// Now, process the name portion ...

	// Separate the words with spaces and capitalize each word.
	name = nameSeparator.ReplaceAllString(name, " ")
	// nolint: staticcheck
	name = strings.Title(name)

	// Remove the spaces to generate a camel case variable name.
	name = strings.ReplaceAll(name, " ", "")

	// Concatenate the names together.
	return ns + name
}

func getFeatureStatus(fs string) (FeatureStatus, error) {
	// nolint: staticcheck
	asTitle := strings.Title(fs)
	switch FeatureStatus(asTitle) {
	case Alpha:
		return Alpha, nil
	case Beta:
		return Beta, nil
	case Stable:
		return Stable, nil
	}
	return "", fmt.Errorf("invalid feature status string: `%s` (stings.Title=`%s`)", fs, asTitle)
}

func generateFeatureStatus(v Variable) string {
	if len(v.FeatureStatus) > 0 {
		fs, err := getFeatureStatus(v.FeatureStatus)
		if err != nil {
			log.Fatal(err)
		}
		return string(fs)
	}

	// If the name begins with the feature status, use it.
	firstPart := strings.Split(v.Name, ".")
	fs, err := getFeatureStatus(firstPart[0])
	if err == nil {
		return string(fs)
	}

	// Default to Alpha
	return string(Alpha)
}

func processHTMLDescription(in string) string {
	// In most cases, the description is a single line in Markdown format.
	// Convert it to HTML with a Markdown parser, this will give us a better looking output.
	return string(mdToHTML([]byte(in)))
}

func processGoDescription(in string, indent int) string {
	if strings.Contains(in, "\n") {
		return lineWrap(in)
	}
	return wordWrap(in, indent)
}

func wordWrap(in string, indent int) string {
	// We use double quotes (") around each line, so replace any double quotes embedded
	// in the string with back ticks (`).
	in = strings.ReplaceAll(in, "\"", "`")

	indentPrefix := strings.Repeat(" ", indent)
	words := strings.Split(in, " ")

	maxLineLength := 80

	out := ""
	line := ""
	for len(words) > 0 {
		// Take the next word.
		word := words[0]
		words = words[1:]

		if indent+len(line)+len(word) > maxLineLength {
			// Need to word wrap - emit the current line.
			out += "\"" + line + " \""
			line = ""

			// Wrap to the start of the next line.
			out += "+\n" + indentPrefix
		}

		// Add the word to the current line.
		if len(line) > 0 {
			line += " "
		}
		line += word
	}

	// Emit the final line
	out += "\"" + line + "\""

	return out
}

func lineWrap(in string) string {
	// We use back tick (`) around the entire string, so replace any back ticks embedded
	// in the string with double quotes (").
	in = strings.ReplaceAll(in, "`", "\"")

	lines := strings.Split(in, "\n")
	out := "`"
	for i, line := range lines {
		out += line
		if i < len(lines)-1 {
			out += "\n"
		}
	}
	out += "`"
	return out
}

func add(x, y int) int {
	return x + y
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
}
