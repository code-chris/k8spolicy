module k8spolicy

go 1.16

require (
	github.com/containerd/containerd v1.5.3
	github.com/deislabs/oras v0.11.1
	github.com/docker/go-units v0.4.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/opencontainers/go-digest v1.0.0
	github.com/opencontainers/image-spec v1.0.1
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.2.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.8.1
	github.com/yargevad/filepathx v1.0.0
	golang.org/x/mod v0.4.2
	helm.sh/helm/v3 v3.1.2
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
