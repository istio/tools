// Copyright 2019 Istio Authors
//
//   Licensed under the Apache License, Version 2.0 (the "License");
//   you may not use this file except in compliance with the License.
//   You may obtain a copy of the License at
//
//       http://www.apache.org/licenses/LICENSE-2.0
//
//   Unless required by applicable law or agreed to in writing, software
//   distributed under the License is distributed on an "AS IS" BASIS,
//   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//   See the License for the specific language governing permissions and
//   limitations under the License.

package generators

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"istio.io/tools/cmd/kubetype-gen/metadata"

	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

type registerGenerator struct {
	generator.DefaultGen
	source  metadata.PackageMetadata
	imports namer.ImportTracker
}

// NewRegisterGenerator creates a new generator for creating k8s style register.go files
func NewRegisterGenerator(source metadata.PackageMetadata) generator.Generator {
	return &registerGenerator{
		DefaultGen: generator.DefaultGen{
			OptionalName: "register",
		},
		source:  source,
		imports: generator.NewImportTracker(),
	}
}

func (g *registerGenerator) Namers(c *generator.Context) namer.NameSystems {
	return NameSystems(g.source.TargetPackage().Path, g.imports)
}

func (g *registerGenerator) PackageConsts(c *generator.Context) []string {
	return []string{
		fmt.Sprintf("GroupName = \"%s\"", g.source.GroupVersion().Group),
	}
}

func (g *registerGenerator) PackageVars(c *generator.Context) []string {
	schemeBuilder := bytes.Buffer{}
	w := bufio.NewWriter(&schemeBuilder)
	sw := generator.NewSnippetWriter(w, c, "$", "$")
	m := map[string]interface{}{
		"NewSchemeBuilder": c.Universe.Function(types.Name{Name: "NewSchemeBuilder", Package: "k8s.io/apimachinery/pkg/runtime"}),
	}
	sw.Do("SchemeBuilder      = $.NewSchemeBuilder|raw$(addKnownTypes)", m)
	w.Flush()
	return []string{
		fmt.Sprintf("SchemeGroupVersion = schema.GroupVersion{Group: GroupName, Version: \"%s\"}", g.source.GroupVersion().Version),
		schemeBuilder.String(),
		"localSchemeBuilder = &SchemeBuilder",
		"AddToScheme        = localSchemeBuilder.AddToScheme",
	}
}

func (g *registerGenerator) Imports(c *generator.Context) []string {
	return g.imports.ImportLines()
}

func (g registerGenerator) Finalize(c *generator.Context, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")
	m := map[string]interface{}{
		"GroupResource":     c.Universe.Type(types.Name{Name: "GroupResource", Package: "k8s.io/apimachinery/pkg/runtime/schema"}),
		"Scheme":            c.Universe.Type(types.Name{Name: "Scheme", Package: "k8s.io/apimachinery/pkg/runtime"}),
		"AddToGroupVersion": c.Universe.Function(types.Name{Name: "AddToGroupVersion", Package: "k8s.io/apimachinery/pkg/apis/meta/v1"}),
		"KubeTypes":         g.source.AllKubeTypes(),
	}
	sw.Do(resourceFuncTemplate, m)
	sw.Do(addKnownTypesFuncTemplate, m)

	return sw.Error()
}

const resourceFuncTemplate = `
func Resource(resource string) $.GroupResource|raw$ {
	return SchemeGroupVersion.WithResource(resource).GroupResource()
}	
`

const addKnownTypesFuncTemplate = `
func addKnownTypes(scheme *$.Scheme|raw$) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		$- range .KubeTypes $
		&$ .Type|raw ${},
		&$ .Type|raw $List{},
		$- end $
	)
	$.AddToGroupVersion|raw$(scheme, SchemeGroupVersion)
	return nil
}
`
