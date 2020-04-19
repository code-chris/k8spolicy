package internal

import (
	"k8spolicy/config"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/yargevad/filepathx"

	"golang.org/x/mod/semver"
)

// DownloadPolicies downloads all configured policies, so they are ready to use.
func DownloadPolicies(skip bool) {
	if skip {
		return
	}

	dir := filepath.Join(config.WorkingDirectory, "policies")
	EnsureDirectory(dir, true)

	configSources := config.Conf.Rules.Additionals

	if Contains(config.Conf.Rules.Presets, "k8s-api-deprecation") {
		configSources = append(configSources, config.RuleSource{URL: "https://github.com/swade1987/deprek8ion", Files: "policies/*.rego"})
	}

	for _, v := range configSources {
		if v.URL == "" {
			// local files
			files, _ := filepathx.Glob(v.Files)
			for _, src := range files {
				CopyFile(src, filepath.Join(dir, filepath.Base(src)))
			}
		} else {
			// from remote
			var url string
			if strings.Contains(v.URL, "github.com") {
				url = v.URL + "/tarball/master"
			} else {
				url = v.URL
			}

			// a tar.gz is assumed
			downloadFile := filepath.Join(dir, "download.tar.gz")

			if err := DownloadFile(downloadFile, url); err != nil {
				log.Fatal(err)
			}

			// extract and copy the files
			downloadDir := filepath.Join(dir, "download")
			EnsureDirectory(downloadDir, false)
			stream, err := os.Open(downloadFile)
			if err != nil {
				log.Fatal(err)
			}

			extractDir := ExtractTarGz(stream, downloadDir)
			stream.Close()
			os.Remove(downloadFile)

			x := filepath.Join(downloadDir, extractDir, v.Files)
			files, _ := filepathx.Glob(x)
			for _, src := range files {
				s, _ := filepath.Rel(filepath.Join(downloadDir, extractDir), src)
				CopyFile(src, filepath.Join(dir, s))
			}

			os.RemoveAll(downloadDir)

			// remove *_test* files
			files, _ = filepathx.Glob(filepath.Join(dir, "**", "*_test*"))
			for _, f := range files {
				os.Remove(f)
			}

			// remove newer version-scoped rule-files
			if config.Conf.TargetVersion != "" {
				r := regexp.MustCompile(`.*(?P<Version>\d\.\d+).*\.rego`)
				files, _ = filepathx.Glob(filepath.Join(dir, "**", "*.rego"))
				for _, f := range files {
					match := r.FindStringSubmatch(filepath.Base(f))
					if len(match) > 0 {
						result := semver.Compare("v"+config.Conf.TargetVersion, "v"+match[1])
						if result == -1 {
							os.Remove(f)
						}
					}
				}
			}
		}
	}
}
