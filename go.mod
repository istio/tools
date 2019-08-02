module istio.io/tools

go 1.12

require (
	cuelang.org/go v0.0.5-0.20190805140455-1372f3e73eee
	fortio.org/fortio v1.1.0
	github.com/client9/gospell v0.0.0-20160306015952-90dfc71015df
	github.com/docker/go-units v0.3.3
	github.com/emicklei/proto v1.6.11
	github.com/getkin/kin-openapi v0.1.1-0.20190507152207-d3180292eead
	github.com/ghodss/yaml v1.0.0
	github.com/gogo/protobuf v1.2.2-0.20190730201129-28a6bbf47e48
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/google/btree v1.0.0 // indirect
	github.com/google/uuid v1.1.1
	github.com/googleapis/gnostic v0.2.0 // indirect
	github.com/gregjones/httpcache v0.0.0-20181110185634-c63ab54fda8f // indirect
	github.com/hashicorp/go-multierror v1.0.0
	github.com/imdario/mergo v0.3.6 // indirect
	github.com/json-iterator/go v1.1.7 // indirect
	github.com/kr/pretty v0.1.0
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/prometheus/client_golang v0.9.3-0.20190127221311-3c4408c8b829
	github.com/shurcooL/httpfs v0.0.0-20190707220628-8d4bc4ba7749
	github.com/shurcooL/sanitized_anchor_name v0.0.0-20170918181015-86672fcb3f95 // indirect
	github.com/shurcooL/vfsgen v0.0.0-20181202132449-6a9ea43bcacd
	github.com/spf13/cobra v0.0.3
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/time v0.0.0-20181108054448-85acf8d2951c // indirect
	golang.org/x/tools v0.0.0-20190614205625-5aca471b1d59
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/russross/blackfriday.v2 v2.0.0-00010101000000-000000000000
	k8s.io/api v0.0.0-20190708094356-59223ed9f6ce
	k8s.io/apiextensions-apiserver v0.0.0-20190708094842-2e94e9ac9840
	k8s.io/apimachinery v0.0.0-20190221084156-01f179d85dbc
	k8s.io/client-go v0.0.0-20190708094436-77c08c6b86df
	k8s.io/gengo v0.0.0-20190128074634-0689ccc1d7d6
	k8s.io/helm v2.12.0+incompatible
	k8s.io/klog v0.3.3 // indirect
)

replace (
	gopkg.in/russross/blackfriday.v2 => github.com/russross/blackfriday/v2 v2.0.1
	k8s.io/api => k8s.io/api v0.0.0-20190708094356-59223ed9f6ce
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190708094842-2e94e9ac9840
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190221084156-01f179d85dbc
	k8s.io/client-go => k8s.io/client-go v0.0.0-20190708094436-77c08c6b86df
)
