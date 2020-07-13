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

// A simple program that consumes a YAML file describing Kubernetes resource annotations and produces as output
// a Go source file providing references to those annotations, and an HTML documentation file describing those
// annotations (for use on istio.io)

package main

import (
	"bytes"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/ghodss/yaml"
	"github.com/spf13/cobra"
)

const (
	outputTemplate = `
// GENERATED FILE -- DO NOT EDIT

package {{ .Package }}

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

// Instance describes a single resource annotation
type Instance struct {
	// The name of the annotation.
	Name string

	// Description of the annotation.
	Description string

	// Hide the existence of this annotation when outputting usage information.
	Hidden bool

	// Mark this annotation as deprecated when generating usage information.
	Deprecated bool

	// The types of resources this annotation applies to.
	Resources []ResourceTypes
}

var (
{{ range .Annotations }}
	{{ .VariableName }} = Instance {
		Name: "{{ .Name }}",
		Description: {{ wordWrap .Description 24 }},
		Hidden: {{ .Hidden }},
		Deprecated: {{ .Deprecated }},
		Resources: []ResourceTypes{
			{{- range .Resources }}
			{{ . }},
			{{- end }}
		},
	}
{{ end }}
)

func AllResourceAnnotations() []*Instance {
	return []*Instance {
		{{- range .Annotations }}
		&{{ .VariableName }},
		{{- end }}
	}
}

func AllResourceTypes() []string {
	return []string {
		{{- range .KnownTypes }}
		"{{ . }}",
		{{- end }}
	}
}`

	htmlOutputTemplate = `---
title: Resource Annotations
description: Resource annotations used by Istio.
location: https://istio.io/docs/reference/config/annotations.html
weight: 60
---
<p>
This page presents the various <a href="https://kubernetes.io/docs/concepts/overview/working-with-objects/annotations/">resource annotations</a> that
Istio supports to control its behavior.
</p>

<table class="annotations">
	<thead>
		<tr>
			<th>Annotation Name</th>
			<th>Resource Types</th>
			<th>Description</th>
		</tr>
	</thead>
	<tbody>
		{{ range .Annotations }}
			{{ if not .Hidden }}
				{{ if .Deprecated }}
					<tr class="deprecated">
				{{ else }}
					<tr>
				{{ end }}
					<td><code>{{ .Name }}</code></td>
					<td>{{ .Resources }}</td>
					<td>{{ .Description }}</td>
				</tr>
			{{ end }}
		{{ end }}
	</tbody>
</table>
`
)

var (
	input      string
	output     string
	htmlOutput string

	nameSeparator = regexp.MustCompile(`[._\-]`)

	rootCmd = cobra.Command{
		Use:   "annotations_prep",
		Short: "Generates a Go source file and HTML file containing annotations.",
		Long:  "Generates a Go source file and HTML file containing annotation definitions based on an input YAML file.",
		Run: func(cmd *cobra.Command, args []string) {
			yamlContent, err := ioutil.ReadFile(input)
			if err != nil {
				log.Fatalf("unable to read input file: %v", err)
			}

			// Unmarshal the file.
			var cfg Configuration
			if err = yaml.Unmarshal(yamlContent, &cfg); err != nil {
				log.Fatalf("error parsing input file: %v", err)
			}

			// Find all the known resource types
			m := make(map[string]bool)
			for _, a := range cfg.Annotations {
				for _, r := range a.Resources {
					m[r] = true
				}
			}
			knownTypes := make([]string, 0, len(m))
			for k := range m {
				knownTypes = append(knownTypes, k)
			}
			sort.Strings(knownTypes)

			// Generate variable names if not provided in the yaml.
			for i := range cfg.Annotations {
				if cfg.Annotations[i].Name == "" {
					log.Fatalf("annotation %d in input file missing name", i)
				}
				if cfg.Annotations[i].VariableName == "" {
					cfg.Annotations[i].VariableName = generateVariableName(cfg.Annotations[i].Name)
				}
			}

			// sort by name
			sort.Slice(cfg.Annotations, func(i, j int) bool {
				return strings.Compare(cfg.Annotations[i].Name, cfg.Annotations[j].Name) < 0
			})

			// Create the output file template.
			t, err := template.New("annoTemplate").Funcs(template.FuncMap{
				"wordWrap": wordWrap, "add": add,
			}).Parse(outputTemplate)
			if err != nil {
				log.Fatalf("failed parsing annotation template: %v", err)
			}

			// Generate the Go source.
			var goSource bytes.Buffer
			if err := t.Execute(&goSource, map[string]interface{}{
				"Package":     getPackage(),
				"KnownTypes":  knownTypes,
				"Annotations": cfg.Annotations,
			}); err != nil {
				log.Fatalf("failed generating output Go source code %s: %v", output, err)
			}

			if err := ioutil.WriteFile(output, goSource.Bytes(), 0666); err != nil {
				log.Fatalf("Failed writing to output file %s: %v", output, err)
			}

			if htmlOutput != "" {
				// Create the HTML output file template.
				t, err = template.New("htmlOutputTemplate").Funcs(template.FuncMap{
					"wordWrap": wordWrap,
				}).Parse(htmlOutputTemplate)
				if err != nil {
					log.Fatalf("failed parsing HTML output template: %v", err)
				}

				// Generate the HTML file.
				var htmlFile bytes.Buffer
				if err := t.Execute(&htmlFile, map[string]interface{}{
					"Package":     getPackage(),
					"Annotations": cfg.Annotations,
				}); err != nil {
					log.Fatalf("failed generating output HTML file %s: %v", output, err)
				}

				if err := ioutil.WriteFile(htmlOutput, htmlFile.Bytes(), 0666); err != nil {
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

	flag.CommandLine.VisitAll(func(gf *flag.Flag) {
		rootCmd.PersistentFlags().AddGoFlag(gf)
	})
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(-1)
	}
}

type AnnotationVariable struct {
	Instance
	VariableName string `json:"variableName"`
}

type Configuration struct {
	Annotations []AnnotationVariable `json:"annotations"`
}

// Instance describes a single resource annotation
type Instance struct {
	// The name of the annotation.
	Name string `json:"name"`

	// Description of the annotation.
	Description string `json:"description"`

	// Hide the existence of this annotation when outputting usage information.
	Hidden bool `json:"hidden"`

	// Mark this annotation as deprecated when generating usage information.
	Deprecated bool `json:"deprecated"`

	// Indicates the types of resources this annotation can be applied to.
	Resources []string `json:"resources"`
}

func getPackage() string {
	path, _ := filepath.Abs(output)
	return filepath.Base(filepath.Dir(path))
}

func generateVariableName(annoName string) string {
	// Split the annotation name to separate the namespace/name portions.
	parts := strings.Split(annoName, "/")
	ns := parts[0]
	name := parts[1]

	// First, process the namespace portion ...

	// Strip .istio.io from the namespace portion of the annotation name.
	ns = strings.TrimSuffix(ns, ".istio.io")

	// Separate the words by spaces and capitalize each word.
	ns = strings.ReplaceAll(ns, ".", " ")
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
	name = strings.Title(name)

	// Remove the spaces to generate a camel case variable name.
	name = strings.ReplaceAll(name, " ", "")

	// Concatenate the names together.
	return ns + name
}

func wordWrap(in string, indent int) string {
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
			out += "+\n"

			// Indent to the start position of the next line.
			for i := 0; i < indent; i++ {
				out += " "
			}
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

func add(x, y int) int {
	return x + y
}
