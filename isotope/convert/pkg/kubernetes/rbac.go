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

package kubernetes

import (
	"fmt"

	"github.com/google/uuid"

	"istio.io/tools/isotope/convert/pkg/graph/svc"
)

func generateRbacPolicy(svc svc.Service, allowAll bool) string {
	tmpl := `
apiVersion: "rbac.istio.io/v1alpha1"
kind: ServiceRole
metadata:
  name: %s
  namespace: %s
spec:
  rules:
  - services: ["%s.%s.*"]
    methods: ["*"]
---
apiVersion: "rbac.istio.io/v1alpha1"
kind: ServiceRoleBinding
metadata:
  name: %s
  namespace: %s
spec:
  subjects:
  - user: "%s"
  roleRef:
    kind: ServiceRole
    name: "%s"
`

	ruleName := uuid.New().String()
	user := ruleName
	if allowAll {
		user = "*"
	}
	ns := ServiceGraphNamespace
	return fmt.Sprintf(tmpl, ruleName, ns, svc.Name, ns, ruleName, ns, user, ruleName)
}

func generateRbacConfig() string {
	tmpl := `
apiVersion: "rbac.istio.io/v1alpha1"
kind: RbacConfig
metadata:
  name: default
spec:
  mode: 'ON_WITH_INCLUSION'
  inclusion:
    namespaces: ["%s"]
`
	return fmt.Sprintf(tmpl, ServiceGraphNamespace)
}
