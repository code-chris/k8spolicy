package internal

import (
	"fmt"
	"k8spolicy/config"
	"k8spolicy/internal/registry"
	"log"
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

// DownloadCharts downloads all configured charts from the repositories and registries
func DownloadCharts() {
	dir := filepath.Join(config.WorkingDirectory, "manifests")
	EnsureDirectory(dir, true)

	downloadFromRepos(dir)
	downloadFromRegistries(dir)
}

func downloadFromRegistries(baseDir string) {
	for _, r := range config.Conf.Helm.Registries {
		registryClient, _ := registry.NewClient()

		ref, err := registry.ParseReference(r.URL + ":" + r.Version)
		if err != nil {
			log.Fatal(err)
		}

		if err = registryClient.PullChart(ref); err != nil {
			log.Fatal(err)
		}

		chart, err := registryClient.LoadChart(ref)
		if err != nil {
			log.Fatal(err)
		}

		if err = chartutil.SaveDir(chart, baseDir); err != nil {
			log.Fatal(err)
		}

		if err = registryClient.RemoveChart(ref); err != nil {
			log.Fatal(err)
		}

		chartDir := filepath.Join(baseDir, chart.Metadata.Name)
		rendered, err := renderChart(chartDir, r.Values)
		if err == nil {
			WriteFile(filepath.Join(baseDir, chart.Metadata.Name+".yaml"), rendered)
		} else {
			fmt.Printf("Skipping %s. Chart cannot be rendered: %e", chart.Metadata.Name, err)
		}

		os.RemoveAll(chartDir)
	}
}

func downloadFromRepos(baseDir string) {
	yaml := `apiVersion: v1
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

	configFile, err := WriteFile(filepath.Join(config.WorkingDirectory, "repositories.yaml"), yaml)
	if err != nil {
		log.Fatal(err)
	}

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
			log.Fatal(err)
		}

		if _, err := repo.DownloadIndexFile(); err != nil {
			log.Fatal(err)
		}

		dest, _, err := dl.DownloadTo(filepath.Join(nameMap[r.URL], r.Chart), r.Version, config.WorkingDirectory)
		if err != nil {
			log.Fatal(err)
		}

		stream, err := os.Open(dest)
		if err != nil {
			log.Fatal(err)
		}

		ExtractTarGz(stream, config.WorkingDirectory)
		stream.Close()
		os.Remove(dest)

		chartDir := filepath.Join(config.WorkingDirectory, r.Chart)
		rendered, err := renderChart(chartDir, r.Values)
		if err == nil {
			WriteFile(filepath.Join(baseDir, r.Chart+".yaml"), rendered)
		} else {
			fmt.Printf("Skipping %s. Chart cannot be rendered: %e", r.Chart, err)
		}

		os.RemoveAll(chartDir)
	}
}

func renderChart(dir string, val []string) (string, error) {
	chart, err := loader.Load(dir)
	if err != nil {
		return "", err
	}

	valueOpts := &values.Options{ValueFiles: val}
	p := getter.All(cli.New())
	vals, err := valueOpts.MergeValues(p)
	if err != nil {
		return "", err
	}

	install := action.NewInstall(&action.Configuration{})
	install.DryRun = true
	install.ClientOnly = true
	install.ReleaseName = "RELEASE-NAME"
	rel, err := install.Run(chart, vals)

	if err != nil {
		return "", err
	}

	return rel.Manifest, nil
}
