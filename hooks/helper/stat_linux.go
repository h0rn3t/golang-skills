package main

import (
	"os"
	"syscall"
)

// changeMeta is the inode and status-change time of fi, which detect a
// rewrite that restores the original bytes and modification time.
func changeMeta(fi os.FileInfo) (ino uint64, ctime int64) {
	if st, ok := fi.Sys().(*syscall.Stat_t); ok {
		return st.Ino, st.Ctim.Nano()
	}
	return 0, 0
}
