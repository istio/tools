module istio.io/tools

go 1.12

replace (
	k8s.io/api => k8s.io/api v0.17.3
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.3
	k8s.io/client-go => k8s.io/client-go v0.17.3
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20181117043124-c2090bec4d9b
)

require (
	cuelang.org/go v0.0.16-0.20200320220106-76252f4b7486
	fortio.org/fortio v1.3.1
	github.com/client9/gospell v0.0.0-20160306015952-90dfc71015df
	github.com/docker/go-units v0.3.3
	github.com/emicklei/proto v1.6.15
	github.com/getkin/kin-openapi v0.1.1-0.20190507152207-d3180292eead
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.3.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/kr/pretty v0.1.0
	github.com/prometheus/client_golang v0.9.3
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/cobra v0.0.4
	github.com/spf13/viper v1.4.0
	golang.org/x/sys v0.0.0-20190924154521-2837fb4f24fe // indirect
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	golang.org/x/tools v0.0.0-20200113154838-30cae5f2fb06
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/yaml.v2 v2.2.8
	istio.io/gogo-genproto v0.0.0-20191009201739-17d570f95998
	k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver v0.0.0-20191011152811-a1d7614a8e0f
	k8s.io/apimachinery v0.17.3
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/gengo v0.0.0-20190822140433-26a664648505
	k8s.io/helm v2.12.0+incompatible
)
