package internal

import (
	"fmt"
	"k8spolicy/config"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// RunConftest executes the conftest binary with the manifests and rules
func RunConftest(skip bool) {
	conftest := downloadConftest(skip)

	yamls, _ := filepath.Glob(filepath.Join(config.WorkingDirectory, "manifests/*.yaml"))
	args := append([]string{"test", "-p", filepath.Join(config.WorkingDirectory, "policies")}, yamls...)

	cmd := exec.Command(conftest, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println(string(output))
		log.Fatal(err)
	}

	fmt.Println(string(output))
	os.Exit(cmd.ProcessState.ExitCode())
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
	version := "0.18.1"
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
