// Copyright Istio Authors
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

package test

import (
	"bufio"
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/google/go-cmp/cmp"

	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	kubeyaml "k8s.io/apimachinery/pkg/util/yaml"
)

const crdFilePath = "generated/kubernetes/customresourcedefinitions.gen.yaml"

func TestCRDwithStatus(t *testing.T) {
	crds := readCRDFile(t)

	for _, c := range crds {
		// status subresource must be enabled.
		if c.Spec.Subresources == nil && c.Spec.Subresources.Status == nil {
			t.Error("status subresource in spec must be enabled.")
		}

		expected := v1beta1.JSONSchemaProps{
			Description: "Status is the test status field.",
			Type:        "object",
			Properties: map[string]v1beta1.JSONSchemaProps{
				"analysis": {
					Type:        "string",
					Format:      "string",
					Description: "Analysis message.",
				},
				"condition": {
					Type:        "string",
					Format:      "string",
					Description: "Current state.",
				},
			},
		}

		got := c.Spec.Validation.OpenAPIV3Schema.Properties["status"]
		if e := cmp.Equal(expected, got); !e {
			t.Errorf("status specs are not equal, expected:\n%v\n, but got:\n%v", expected, got)
		}

	}
}

func readCRDFile(t *testing.T) []*v1beta1.CustomResourceDefinition {
	r, err := os.Open(crdFilePath)
	if err != nil {
		t.Fatal(err)
	}

	f := bufio.NewReader(r)

	decoder := kubeyaml.NewYAMLReader(f)
	crds := make([]*v1beta1.CustomResourceDefinition, 0)
	for {
		doc, err := decoder.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("unable to read yaml document: %v", err)
			break
		}
		chunk := bytes.TrimSpace(doc)

		if len(chunk) == 0 {
			break
		}

		crd := v1beta1.CustomResourceDefinition{}
		if err = yaml.Unmarshal(chunk, &crd); err != nil {
			t.Fatal(err)
		}

		crds = append(crds, &crd)

	}
	return crds
}
