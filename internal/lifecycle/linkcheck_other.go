//go:build !windows

package lifecycle

import "os"

// isLinked reports whether the entry is a symbolic link.
func isLinked(info os.FileInfo) bool {
	return info.Mode()&os.ModeSymlink != 0
}
