package testhelpers

import (
	"log"
	"os"
	"path/filepath"
)

// ModuleRootDir finds the location of the root directory of this respository.
// Note: typically Go tests can assume a fixed directory location, but some
// tests/examples in this repository get copied and can run from multiple
// directories, requiring the use of this function.
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
