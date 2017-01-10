// Copyright 2016 The Minimal Configuration Manager Authors
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

// +build !windows

package system

import (
	"errors"
	"os"
	"syscall"
)

// LocalRoot is Local's root filesystem path.
const LocalRoot = "/"

// OwnerInfo attempts to retrieve a file's uid and gid from info.Sys().
func (Local) OwnerInfo(info os.FileInfo) (UID, GID, error) {
	st, ok := info.Sys().(*syscall.Stat_t)
	if !ok {
		return 0, 0, errors.New("file info has no uid/gid fields")
	}
	return UID(st.Uid), GID(st.Gid), nil
}
