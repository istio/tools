module istio.io/tools

go 1.12

replace (
	// For license
	github.com/pelletier/go-buffruneio => github.com/pelletier/go-buffruneio v0.3.0
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
	github.com/gogo/protobuf v1.3.1
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/golang/protobuf v1.4.2
	github.com/google/go-cmp v0.5.0
	github.com/hhatto/gorst v0.0.0-20181029133204-ca9f730cac5b // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jdkato/prose v1.1.0 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20180830205328-81db2a75821e // indirect
	github.com/kr/pretty v0.1.0
	github.com/montanaflynn/stats v0.0.0-20180911141734-db72e6cae808 // indirect
	github.com/neurosnap/sentences v1.0.6 // indirect
	github.com/pelletier/go-buffruneio v0.2.0 // indirect
	github.com/russross/blackfriday/v2 v2.0.1
	github.com/shogo82148/go-shuffle v0.0.0-20180218125048-27e6095f230d // indirect
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.4.0
	github.com/src-d/gcfg v1.4.0 // indirect
	github.com/xanzy/ssh-agent v0.2.0 // indirect
	golang.org/x/tools v0.0.0-20200115044656-831fdb1e1868
	gonum.org/v1/netlib v0.0.0-20191031114514-eccb95939662 // indirect
	gopkg.in/neurosnap/sentences.v1 v1.0.6 // indirect
	gopkg.in/russross/blackfriday.v2 v2.0.0 // indirect
	gopkg.in/src-d/go-billy-siva.v4 v4.2.2 // indirect
	gopkg.in/src-d/go-billy.v4 v4.3.0 // indirect
	gopkg.in/src-d/go-git-fixtures.v3 v3.5.0 // indirect
	gopkg.in/src-d/go-git.v4 v4.7.0 // indirect
	gopkg.in/src-d/go-license-detector.v2 v2.0.0-20180510072912-da552ecf050b
	gopkg.in/src-d/go-siva.v1 v1.3.0 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v2 v2.3.0
	istio.io/api v0.0.0-20200722144311-7e311b6ce256
	istio.io/gogo-genproto v0.0.0-20200720193312-b523a30fe746
	istio.io/istio v0.0.0-20200722232529-c37d4187c0e6
	k8s.io/apiextensions-apiserver v0.18.3
	k8s.io/apimachinery v0.18.3
	k8s.io/gengo v0.0.0-20190822140433-26a664648505
	k8s.io/utils v0.0.0-20200414100711-2df71ebbae66
	sigs.k8s.io/controller-tools v0.3.0
)
