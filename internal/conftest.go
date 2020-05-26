package internal

import (
	"fmt"
	"k8spolicy/config"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/yargevad/filepathx"
	"golang.org/x/mod/semver"
)

// RunConftest executes the conftest binary with the manifests and rules
func RunConftest(skip bool) {
	conftest := downloadConftest(skip)
	filterRuleFiles()
	copyStandaloneYamlFiles()

	yamls, _ := filepathx.Glob(filepath.Join(config.WorkingDirectory, "manifests/**/*.yaml"))
	args := append([]string{"test", "-p", filepath.Join(config.WorkingDirectory, "currentPolicies")}, yamls...)

	cmd := exec.Command(conftest, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		log.Fatal(err)
	}

	fmt.Println(string(output))
	os.Exit(cmd.ProcessState.ExitCode())
}

func filterRuleFiles() {
	policies, _ := filepathx.Glob(filepath.Join(config.WorkingDirectory, "policies/**/*.rego"))
	dir := filepath.Join(config.WorkingDirectory, "currentPolicies")
	EnsureDirectory(dir, true)
	r := regexp.MustCompile(`.*(?P<Version>\d\.\d+).*\.rego`)

	for _, f := range policies {
		if (strings.Contains(f, "k8s-api-deprecation") && !Contains(config.Conf.Rules.Presets, "k8s-api-deprecation")) ||
			(strings.Contains(f, "k8s-security") && !Contains(config.Conf.Rules.Presets, "k8s-security")) {
			continue
		}

		match := r.FindStringSubmatch(filepath.Base(f))
		if len(match) > 0 && config.Conf.TargetVersion != "" {
			result := semver.Compare("v"+config.Conf.TargetVersion, "v"+match[1])
			if result != -1 {
				s, _ := filepath.Rel(filepath.Join(config.WorkingDirectory, "policies"), f)
				CopyFile(f, filepath.Join(dir, s))
			}
		} else {
			s, _ := filepath.Rel(filepath.Join(config.WorkingDirectory, "policies"), f)
			CopyFile(f, filepath.Join(dir, s))
		}
	}
}

func downloadConftest(skip bool) string {
	conftest := filepath.Join(config.WorkingDirectory, "conftest")

	if skip {
		return conftest
	}

	if _, err := os.Stat(conftest); err == nil {
		return conftest
	}

	arch := runtime.GOOS
	version := "0.18.2"
	fmt.Println("Downloading conftest " + version + "...")
	url := "https://github.com/instrumenta/conftest/releases/download/v" + version + "/conftest_" + version + "_" + arch + "_x86_64.tar.gz"

	downloadFile := filepath.Join(config.WorkingDirectory, "conftest.tar.gz")
	if err := DownloadFile(downloadFile, url); err != nil {
		log.Fatal(err)
	}

	stream, err := os.Open(downloadFile)
	if err != nil {
		log.Fatal(err)
	}

	ExtractTarGz(stream, config.WorkingDirectory)
	stream.Close()
	os.Remove(downloadFile)
	os.Chmod(conftest, 0755)
	return conftest
}

func copyStandaloneYamlFiles() {
	basePath := filepath.Join(config.WorkingDirectory, "manifests")

	for _, glob := range config.Conf.Files {
		yamls, _ := filepath.Glob(glob)

		for _, yaml := range yamls {
			dest := filepath.Join(basePath, yaml)
			CopyFile(yaml, dest)
		}
	}
}
