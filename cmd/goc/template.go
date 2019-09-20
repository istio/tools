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
	"bytes"
	"text/template"
)

var (
	tmpl     = template.Must(template.New("reg").Parse(tmplText))
	tmplText = `//
// Generated Code, DO NOT EDIT!
//

package {{.PackageName}}

import (
	"sync/atomic"

	_cover_ "istio.io/pkg/cover"
)

func init() {
{{range $name, $var := .ContextVars }}
	_cover_.GetRegistry().Register(
		len({{$var}}.Count),
		"{{$name}}",
		func(o []uint32) {
			l := len({{$var}}.Pos)
			for i := 0; i < l; i++ {
				o[i] = {{$var}}.Pos[i]
			}
		},
		func(o []uint16) {
			l := len({{$var}}.NumStmt)
			for i := 0; i < l; i++ {
				o[i] = {{$var}}.NumStmt[i]
			}
		},
		func(o []uint32) {
			l := len({{$var}}.Count)
			for i := 0; i < l; i++ {
				o[i] = atomic.LoadUint32(&{{$var}}.Count[i])
			}
		},
		func() {
			l := len({{$var}}.Count)
			for i := 0; i < l; i++ {
				atomic.StoreUint32(&{{$var}}.Count[i], 0)
			}
		})
{{end}}
}
`
)

func generateRegistrationCode(pkgName string, contextVars map[string]string) (string, error) {
	var b bytes.Buffer
	err := tmpl.Execute(&b, &struct {
		PackageName string
		ContextVars map[string]string
	}{
		PackageName: pkgName,
		ContextVars: contextVars,
	})
	if err != nil {
		return "", err
	}

	return b.String(), nil
}
