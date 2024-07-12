// Copyright 2023 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
