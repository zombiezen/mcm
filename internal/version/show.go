// Copyright 2017 The Minimal Configuration Manager Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package version

import (
	"fmt"
	"os"
	"path/filepath"
)

// Show prints the executable's version information to stderr.
func Show() {
	exeName := filepath.Base(os.Args[0])
	switch {
	case Label != "":
		fmt.Fprintf(os.Stderr, "%s: version %s\n", exeName, Label)
	case SCMStatus == "Modified":
		fmt.Fprintf(os.Stderr, "%s: built from %s with local modifications\n", exeName, SCMRevision)
	default:
		fmt.Fprintf(os.Stderr, "%s: built from %s\n", exeName, SCMRevision)
	}
}
