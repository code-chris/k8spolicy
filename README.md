
# K8sPolicy

This cli-tool helps you to run rego-policies against your kubernetes yaml-files. Helm-Charts are supported as well. [Conftest](https://github.com/instrumenta/conftest) is used under the hood.

## Installation

Download the appropriate binary from the [releases](https://github.com/code-chris/k8spolicy/releases) page.

## Usage
```
Run's all configured rules against the manifests to test

Usage:
  k8spolicy test [flags]

Flags:
  -h, --help                     help for test
      --skip-conftest-download   Do not download the conftest binary
      --skip-policy-download     Do not download the policy files

Global Flags:
      --config string   config file (default is .k8spolicy.yaml)
```

A configuration file is required. See below for details.

### ENV

The environment variables `K8SPOLICY_SKIP_POLICY_DOWNLOAD` and `K8SPOLICY_SKIP_CONFTEST_DOWNLOAD` can be set to `true` as cli-flag replacement.

## Configuration
```
rules:
  presets:
    - k8s-api-deprecation
    - k8s-security
  additionals:
    - files: path/to/my/policies/*.rego
    - url: https://github.com/instrumenta/policies
      files: kubernetes/**/*.rego
targetVersion: 1.17
helm:
  repositories:
    - url: https://charts.bitnami.com/bitnami
      chart: nginx-ingress-controller
      version: 5.3.13
      values:
        - charts/nginx-chart.yaml
  registries:
    - url: registry.mycompany.com/charts
      version: 1.0.0
      values:
        - charts/myawesome-chart.yaml
files:
  - additional/manifest/files/*.yaml
```
All filesystem paths are relative to the execution directory of the cli-tool.

#### rules.presets
Use this array to automatically include one of these presets:
| Name                 | URL                                     |
| -------------------- | --------------------------------------- |
| k8s-api-deprecation  | https://github.com/swade1987/deprek8ion |
| k8s-security         | https://github.com/instrumenta/policies |

#### rules.additionals
If only files are specified, this is determined as local path. If a url is also given, then the files are downloaded. It is assumed, that the download-url resolves to a `tar.gz` file. In case of a `github.com` (as above) the current master-tarball is used. Only the given filepath is used from the downloaded files.

#### targetVersion (optional)
If a rego-file includes a kubernetes-version (this regex is used: `.*(\d\.\d+).*\.rego`) you can exclude those files which have a greater version.

#### helm.repositories
A list of charts from repositories to validate. `url` and `chart` are mandatory, `version` and `values` are optional.

#### helm.registries
A list of charts from OCI-Registries to validate. `url` and `version` are mandatory, `values` are optional.

#### files
A list of any other local yaml-files which should be validated.


You have to use a preset or specify any additional rules. One of the `helm.repositories`, `helm.registries` or `files` is also mandatory.


## Docker

There's a docker-image for usage in CI-Environemnts for example.
```
docker pull ckotzbauer/k8spolicy
```

The image is pre-populated with the `conftest` binary and the two presets. The download of both is disabled by default.


## Roadmap

- Support helm-registries with authentication.
- Filter results with a regex.
