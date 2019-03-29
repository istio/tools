// Copyright 2018 Istio Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this currentFile except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package graphviz converts service graphs into Graphviz DOT language.
package graphviz

import (
	"bytes"
	"fmt"
	"text/template"

	"istio.io/tools/isotope/convert/pkg/graph"
	"istio.io/tools/isotope/convert/pkg/graph/script"
	"istio.io/tools/isotope/convert/pkg/graph/svc"
)

// ServiceGraphToDotLanguage converts a ServiceGraph to a Graphviz DOT language
// string.
func ServiceGraphToDotLanguage(
	serviceGraph graph.ServiceGraph) (string, error) {
	graph, err := ServiceGraphToGraph(serviceGraph)
	if err != nil {
		return "", err
	}
	dotLang, err := GraphToDotLanguage(graph)
	if err != nil {
		return "", err
	}
	return dotLang, nil
}

// GraphToDotLanguage converts a graphviz graph to a Graphviz DOT language
// string via a template.
func GraphToDotLanguage(g Graph) (string, error) {
	tmpl, err := template.New("digraph").Parse(graphvizTemplate)
	if err != nil {
		return "", err
	}

	var b bytes.Buffer
	err = tmpl.Execute(&b, g)
	if err != nil {
		return "", err
	}
	return b.String(), nil
}

// ServiceGraphToGraph converts a service graph to a graphviz graph.
func ServiceGraphToGraph(sg graph.ServiceGraph) (Graph, error) {
	nodes := make([]Node, 0, len(sg.Services))
	edges := make([]Edge, 0, len(sg.Services))
	for _, service := range sg.Services {
		node, connections, err := toGraphvizNode(service)
		if err != nil {
			return Graph{}, err
		}
		nodes = append(nodes, node)
		for _, connection := range connections {
			edges = append(edges, connection)
		}
	}
	return Graph{
		Nodes: nodes,
		Edges: edges,
	}, nil
}

// Graph represents a Graphviz graph.
type Graph struct {
	Nodes []Node
	Edges []Edge
}

// Node represents a node in the Graphviz graph.
type Node struct {
	Name         string
	Type         string
	ErrorRate    string
	ResponseSize string
	Steps        [][]string
}

// Edge represents a directed edge in the Graphviz graph.
type Edge struct {
	From      string
	To        string
	StepIndex int
}

const graphvizTemplate = `digraph {
  node [
    fontsize = "16"
    fontname = "courier"
    shape = plaintext
  ];

  {{ range .Nodes -}}
  "{{ .Name }}" [label=<
<TABLE BORDER="0" CELLBORDER="1" CELLSPACING="0">
  <TR><TD><B>{{ .Name }}</B><BR />Type: {{ .Type }}<BR />Err: {{ .ErrorRate }}</TD></TR>
  {{- range $i, $cmds := .Steps }}
  <TR><TD PORT="{{ $i }}">
  {{- range $j, $cmd := $cmds -}}
    {{- if $j -}}<BR />{{- end -}}
    {{- $cmd -}}
  {{- end -}}
  </TD></TR>
  {{- end }}
</TABLE>>];

  {{ end }}

  {{- range .Edges }}
  "{{ .From -}}":{{- .StepIndex }} -> "{{ .To }}"
  {{- end }}
}
`

func getEdgesFromExe(
	exe script.Command, idx int, fromServiceName string) (edges []Edge) {
	switch cmd := exe.(type) {
	case script.ConcurrentCommand:
		for _, subCmd := range cmd {
			subEdges := getEdgesFromExe(subCmd, idx, fromServiceName)
			for _, e := range subEdges {
				edges = append(edges, e)
			}
		}
	case script.RequestCommand:
		e := Edge{
			From:      fromServiceName,
			To:        cmd.ServiceName,
			StepIndex: idx,
		}
		edges = append(edges, e)
	}
	return
}

func toGraphvizNode(service svc.Service) (Node, []Edge, error) {
	steps := make([][]string, 0, len(service.Script))
	edges := make([]Edge, 0, len(service.Script))
	for idx, exe := range service.Script {
		step, err := executableToStringSlice(exe)
		if err != nil {
			return Node{}, nil, err
		}
		steps = append(steps, step)

		stepEdges := getEdgesFromExe(exe, idx, service.Name)
		for _, e := range stepEdges {
			edges = append(edges, e)
		}
	}
	n := Node{
		Name:         service.Name,
		Type:         service.Type.String(),
		ErrorRate:    service.ErrorRate.String(),
		ResponseSize: service.ResponseSize.String(),
		Steps:        steps,
	}
	return n, edges, nil
}

func nonConcurrentCommandToString(exe script.Command) (string, error) {
	switch cmd := exe.(type) {
	case script.SleepCommand:
		return fmt.Sprintf("SLEEP %s", cmd), nil
	case script.RequestCommand:
		return fmt.Sprintf(
			"CALL \"%s\" %s",
			cmd.ServiceName, cmd.Size.String()), nil
	default:
		return "", fmt.Errorf("unexpected type of executable %T", exe)
	}
}

func executableToStringSlice(exe script.Command) ([]string, error) {
	slice := make([]string, 0, 1)
	appendNonConcurrentExe := func(exe script.Command) error {
		s, err := nonConcurrentCommandToString(exe)
		if err != nil {
			return err
		}
		slice = append(slice, s)
		return nil
	}

	switch cmd := exe.(type) {
	case script.SleepCommand:
		if err := appendNonConcurrentExe(exe); err != nil {
			return nil, err
		}
	case script.RequestCommand:
		if err := appendNonConcurrentExe(exe); err != nil {
			return nil, err
		}
	case script.ConcurrentCommand:
		for _, exe := range cmd {
			if err := appendNonConcurrentExe(exe); err != nil {
				return nil, err
			}
		}
	default:
		return nil, fmt.Errorf("unexpected type of executable %T", exe)
	}
	return slice, nil
}
