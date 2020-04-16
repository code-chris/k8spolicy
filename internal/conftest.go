package internal

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func RunConftest() {
	conftest := downloadConftest()

	yamls, _ := filepath.Glob("/tmp/k8spolicy/manifests/*.yaml")
	args := append([]string{"test", "-p", "/tmp/k8spolicy/policies"}, yamls...)

	cmd := exec.Command(conftest, args...)
	output, _ := cmd.CombinedOutput()
	fmt.Println(string(output))
	os.Exit(cmd.ProcessState.ExitCode())
}

func downloadConftest() string {
	dir := filepath.Join(os.TempDir(), "k8spolicy")
	conftest := filepath.Join(dir, "conftest")

	if _, err := os.Stat(conftest); err == nil {
		return conftest
	}

	arch := runtime.GOOS
	version := "0.18.1"
	fmt.Println("Downloading conftest " + version + "...")
	url := "https://github.com/instrumenta/conftest/releases/download/v" + version + "/conftest_" + version + "_" + arch + "_x86_64.tar.gz"

	downloadFile := filepath.Join(dir, "conftest.tar.gz")
	err := DownloadFile(downloadFile, url)

	if err != nil {
		panic(err)
	}

	stream, _ := os.Open(downloadFile)
	ExtractTarGz(stream, dir)
	stream.Close()
	os.Remove(downloadFile)
	os.Chmod(conftest, 0755)
	return conftest
}
