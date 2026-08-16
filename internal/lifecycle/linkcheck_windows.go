//go:build windows

package lifecycle

import (
	"os"
	"syscall"
)

// isLinked reports whether the entry is a symlink or any other reparse point,
// such as a junction or mount point. Go's Lstat only reports IO_REPARSE_TAG_SYMLINK
// entries as ModeSymlink; junction points are reported as directories and must
// be detected through the raw Win32 file attributes.
func isLinked(info os.FileInfo) bool {
	if info.Mode()&os.ModeSymlink != 0 {
		return true
	}
	data, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return false
	}
	return data.FileAttributes&syscall.FILE_ATTRIBUTE_REPARSE_POINT != 0
}
