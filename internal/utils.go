package internal

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

// CopyFile creates a file at the given destination and writes all bytes from src
func CopyFile(src string, dest string) {
	from, err := os.Open(src)
	if err != nil {
		log.Fatal(err)
	}
	defer from.Close()

	EnsureDirectory(filepath.Dir(dest), false)
	to, err := os.OpenFile(dest, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer to.Close()

	if _, err = io.Copy(to, from); err != nil {
		log.Fatal(err)
	}
}

// Contains returns true, if the given array contains the given string
func Contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}

// DownloadFile will download a url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func DownloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// ExtractTarGz extracts all files from the stream to the given basePath
func ExtractTarGz(gzipStream io.Reader, basePath string) string {
	uncompressedStream, err := gzip.NewReader(gzipStream)
	if err != nil {
		log.Fatal("ExtractTarGz: NewReader failed")
	}

	tarReader := tar.NewReader(uncompressedStream)
	var rootDir string

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		} else if err != nil {
			log.Fatalf("ExtractTarGz: Next() failed: %s", err.Error())
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if rootDir == "" {
				rootDir = header.Name
			}

			if err := os.Mkdir(filepath.Join(basePath, header.Name), 0755); err != nil {
				log.Fatalf("ExtractTarGz: Mkdir() failed: %s", err.Error())
			}
		case tar.TypeReg:
			EnsureDirectory(filepath.Dir(filepath.Join(basePath, header.Name)), false)
			outFile, err := os.Create(filepath.Join(basePath, header.Name))
			if err != nil {
				log.Fatalf("ExtractTarGz: Create() failed: %s", err.Error())
			}

			if _, err := io.Copy(outFile, tarReader); err != nil {
				log.Fatalf("ExtractTarGz: Copy() failed: %s", err.Error())
			}

			outFile.Close()

		default:
			/*log.Fatalf(
			"ExtractTarGz: uknown type: %s in %s",
			header.Typeflag,
			header.Name)*/
		}
	}

	return rootDir
}

// EnsureDirectory creates the given path and removes it first if the clear flag is set
func EnsureDirectory(path string, clear bool) {
	if clear == true {
		_ = os.RemoveAll(path)
		_ = os.MkdirAll(path, 0755)
	} else if _, err := os.Stat(path); err != nil {
		_ = os.MkdirAll(path, 0755)
	}
}

// WriteFile creates a file at the given path and writes the whole string to it.
func WriteFile(path string, s string) (string, error) {
	f, err := os.Create(path)
	if err != nil {
		log.Fatal(err)
	}

	if _, err = f.WriteString(s); err != nil {
		f.Close()
		log.Fatal(err)
	}

	return path, nil
}
