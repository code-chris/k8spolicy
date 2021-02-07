module k8spolicy

go 1.15

require (
	github.com/containerd/containerd v1.4.3
	github.com/deislabs/oras v0.10.0
	github.com/docker/go-units v0.4.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/mitchellh/mapstructure v1.2.2 // indirect
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/pelletier/go-toml v1.7.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/yargevad/filepathx v0.0.0-20161019152617-907099cb5a62
	golang.org/x/mod v0.4.1
	gopkg.in/ini.v1 v1.55.0 // indirect
	helm.sh/helm/v3 v3.1.2
	rsc.io/letsencrypt v0.0.3 // indirect
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
