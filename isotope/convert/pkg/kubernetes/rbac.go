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
