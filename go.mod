module github.com/kong/kubernetes-ingress-controller

go 1.13

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/eapache/channels v1.1.0
	github.com/fatih/color v1.7.0
	github.com/golang/glog v0.0.0-20160126235308-23def4e6c14b
	github.com/hashicorp/go-uuid v1.0.1
	github.com/hbagdi/deck v0.7.1-0.20191223190449-3d9f90945d9d
	github.com/hbagdi/go-kong v0.10.0
	github.com/mattn/go-colorable v0.1.2 // indirect
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkg/errors v0.9.1
	github.com/prometheus/client_golang v1.11.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/stretchr/testify v1.7.0
	github.com/tidwall/gjson v1.2.1
	github.com/tidwall/match v1.0.1 // indirect
	github.com/tidwall/pretty v0.0.0-20190325153808-1166b9ac2b65 // indirect
	gopkg.in/go-playground/assert.v1 v1.2.1 // indirect
	gopkg.in/go-playground/pool.v3 v3.1.1
	k8s.io/api v0.20.7
	k8s.io/apimachinery v0.20.7
	k8s.io/client-go v0.20.7
	knative.dev/serving v0.24.0
)
