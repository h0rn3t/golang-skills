//go:build !linux && !darwin

package main

import "os"

// changeMeta has no status-change time here: content and modification time
// alone detect a rewrite.
func changeMeta(os.FileInfo) (ino uint64, ctime int64) {
	return 0, 0
}
