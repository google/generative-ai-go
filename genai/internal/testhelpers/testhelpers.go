package testhelpers

import (
	"log"
	"os"
	"path/filepath"
)

// ModuleRootDir finds the location of the root directory of this respository.
// Note: typically Go tests can assume a fixed directory location, but this
// particular file gets copied and can run from multiple directories (see
// the generate directive above).
func ModuleRootDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Getcwd:", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			log.Fatal("unable to find")
		}
		dir = parentDir
	}
}
