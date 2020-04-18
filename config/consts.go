package config

import (
	"os"
	"path/filepath"
)

var (
	// WorkingDirectory specifies the root directory used for all operations of this program
	WorkingDirectory = filepath.Join(os.TempDir(), "k8spolicy")
)
