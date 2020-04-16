package internal

import (
	"fmt"
	"k8spolicy/config"
	"k8spolicy/internal/registry"
	"os"
	"path/filepath"
	"regexp"

	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/chart/loader"
	"helm.sh/helm/v3/pkg/chartutil"
	"helm.sh/helm/v3/pkg/cli"
	"helm.sh/helm/v3/pkg/cli/values"
	"helm.sh/helm/v3/pkg/downloader"
	"helm.sh/helm/v3/pkg/getter"
	"helm.sh/helm/v3/pkg/repo"
)

func DownloadCharts() {
	downloadFromRepos()
	downloadFromRegistries()
}

func downloadFromRegistries() {
	for _, r := range config.Conf.Helm.Registries {
		registryClient, _ := registry.NewClient()

		ref, err := registry.ParseReference(r.URL + ":" + r.Version)
		if err != nil {
			panic(err)
		}

		err = registryClient.PullChart(ref)
		if err != nil {
			panic(err)
		}

		chart, err := registryClient.LoadChart(ref)
		if err != nil {
			panic(err)
		}

		EnsureDirectory("/tmp/k8spolicy/manifests", false)
		err = chartutil.SaveDir(chart, "/tmp/k8spolicy/manifests")
		if err != nil {
			panic(err)
		}

		err = registryClient.RemoveChart(ref)
		if err != nil {
			panic(err)
		}

		chartDir := filepath.Join("/tmp/k8spolicy/manifests", chart.Metadata.Name)
		rendered := renderChart(chartDir, r.Values)
		WriteFile("/tmp/k8spolicy/manifests/"+chart.Metadata.Name+".yaml", rendered)
		os.RemoveAll(chartDir)
	}
}

func downloadFromRepos() {
	yaml := `
apiVersion: v1
repositories:
`

	reg, _ := regexp.Compile("[^a-zA-Z0-9]+")
	var nameMap = make(map[string]string)

	for _, r := range config.Conf.Helm.Repositories {
		cleaned := reg.ReplaceAllString(r.URL, "")
		yaml += fmt.Sprintf("- name: %s\n", cleaned)
		yaml += fmt.Sprintf("  url: %s\n", r.URL)
		nameMap[r.URL] = cleaned
	}

	configFile, _ := WriteFile("/tmp/k8spolicy/repositories.yaml", yaml)
	settings := cli.New()

	dl := downloader.ChartDownloader{
		Out:              os.Stdout,
		Getters:          getter.All(settings),
		RepositoryConfig: configFile,
		RepositoryCache:  settings.RepositoryCache,
	}

	for _, r := range config.Conf.Helm.Repositories {
		repo, err := repo.NewChartRepository(&repo.Entry{URL: r.URL, Name: nameMap[r.URL]}, getter.All(settings))
		if err != nil {
			panic(err)
		}

		_, err = repo.DownloadIndexFile()
		if err != nil {
			panic(err)
		}

		dest, _, err := dl.DownloadTo(nameMap[r.URL]+"/"+r.Chart, r.Version, "/tmp/k8spolicy")
		if err != nil {
			panic(err)
		}

		stream, _ := os.Open(dest)
		ExtractTarGz(stream, "/tmp/k8spolicy")
		stream.Close()
		os.Remove(dest)

		EnsureDirectory("/tmp/k8spolicy/manifests", false)
		chartDir := filepath.Join("/tmp/k8spolicy", r.Chart)
		rendered := renderChart(chartDir, r.Values)
		WriteFile("/tmp/k8spolicy/manifests/"+r.Chart+".yaml", rendered)
		os.RemoveAll(chartDir)
	}
}

func renderChart(dir string, val []string) string {
	chart, err := loader.Load(dir)
	if err != nil {
		panic(err)
	}

	valueOpts := &values.Options{ValueFiles: val}
	p := getter.All(cli.New())
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		panic(err)
	}

	install := action.NewInstall(&action.Configuration{})
	install.DryRun = true
	install.ClientOnly = true
	install.ReleaseName = "RELEASE-NAME"
	rel, err := install.Run(chart, vals)
	return rel.Manifest
}
