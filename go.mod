module istio.io/tools

go 1.12

replace (
	// For license
	github.com/pelletier/go-buffruneio => github.com/pelletier/go-buffruneio v0.3.0
	k8s.io/api => k8s.io/api v0.0.0-20190817221950-ebce17126a01
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20191011152811-a1d7614a8e0f
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817221809-bf4de9df677c
	k8s.io/client-go => k8s.io/client-go v10.0.0+incompatible
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20181117043124-c2090bec4d9b
)

require (
	cuelang.org/go v0.0.16-0.20200320220106-76252f4b7486
	github.com/alcortesm/tgz v0.0.0-20161220082320-9c5fe88206d7 // indirect
	github.com/anmitsu/go-shlex v0.0.0-20161002113705-648efa622239 // indirect
	github.com/client9/gospell v0.0.0-20160306015952-90dfc71015df
	github.com/dgryski/go-metro v0.0.0-20180109044635-280f6062b5bc // indirect
	github.com/dgryski/go-minhash v0.0.0-20170608043002-7fe510aff544 // indirect
	github.com/dgryski/go-spooky v0.0.0-20170606183049-ed3d087f40e2 // indirect
	github.com/ekzhu/minhash-lsh v0.0.0-20171225071031-5c06ee8586a1 // indirect
	github.com/emicklei/proto v1.6.15
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/flynn/go-shlex v0.0.0-20150515145356-3f9db97f8568 // indirect
	github.com/getkin/kin-openapi v0.1.1-0.20190507152207-d3180292eead
	github.com/ghodss/yaml v1.0.0
	github.com/gliderlabs/ssh v0.2.2 // indirect
	github.com/gogo/protobuf v1.3.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.3.2
	github.com/googleapis/gnostic v0.3.1 // indirect
	github.com/gregjones/httpcache v0.0.0-20190611155906-901d90724c79 // indirect
	github.com/hhatto/gorst v0.0.0-20181029133204-ca9f730cac5b // indirect
	github.com/imdario/mergo v0.3.8 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jdkato/prose v1.1.0 // indirect
	github.com/json-iterator/go v1.1.7 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20180830205328-81db2a75821e // indirect
	github.com/kr/pretty v0.1.0
	github.com/modern-go/reflect2 v1.0.1 // indirect
	github.com/montanaflynn/stats v0.0.0-20180911141734-db72e6cae808 // indirect
	github.com/neurosnap/sentences v1.0.6 // indirect
	github.com/pelletier/go-buffruneio v0.2.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/sergi/go-diff v1.0.0 // indirect
	github.com/shogo82148/go-shuffle v0.0.0-20180218125048-27e6095f230d // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/cobra v0.0.4
	github.com/spf13/viper v1.4.0
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/xanzy/ssh-agent v0.2.0 // indirect
	golang.org/x/net v0.0.0-20191004110552-13f9640d40b9 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45 // indirect
	golang.org/x/sys v0.0.0-20190924154521-2837fb4f24fe // indirect
	golang.org/x/time v0.0.0-20190921001708-c4c64cad1fd0 // indirect
	golang.org/x/tools v0.0.0-20200113154838-30cae5f2fb06
	gonum.org/v1/netlib v0.0.0-20191031114514-eccb95939662 // indirect
	google.golang.org/appengine v1.6.5 // indirect
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/neurosnap/sentences.v1 v1.0.6 // indirect
	gopkg.in/russross/blackfriday.v2 v2.0.0 // indirect
	gopkg.in/src-d/go-billy-siva.v4 v4.2.2 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.0 // indirect
	gopkg.in/src-d/go-git-fixtures.v3 v3.5.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.7.0 // indirect
	gopkg.in/src-d/go-license-detector.v2 v2.0.0-20180510072912-da552ecf050b
	gopkg.in/src-d/go-siva.v1 v1.3.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.2.4
	istio.io/gogo-genproto v0.0.0-20191009201739-17d570f95998
	k8s.io/api v0.0.0-20191004120003-3a12735a829a
	k8s.io/apiextensions-apiserver v0.0.0-20191011152811-a1d7614a8e0f
	k8s.io/apimachinery v0.0.0-20191004115801-a2eda9f80ab8
	k8s.io/client-go v0.0.0-20191016111102-bec269661e48
	k8s.io/gengo v0.0.0-20190822140433-26a664648505
	k8s.io/helm v2.12.0+incompatible
	k8s.io/klog v0.4.0 // indirect
)
