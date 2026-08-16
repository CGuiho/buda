//go:build windows

package selfmanage

import "fmt"

// Windows keeps the running process immutable. The stable launcher owns the
// process boundary, so lifecycle operations replace an inactive versioned
// payload instead of copying over the executable that is currently running.
func replaceExecutable(executable, candidate, backup, targetVersion, checksum, wiki string, _ VerifyFunc) (bool, error) {
	return false, fmt.Errorf("synchronous replacement of the running Windows executable is unavailable; install Buda through the stable launcher")
}

func RemoveExecutable(path string) (bool, error) {
	return false, fmt.Errorf("synchronous removal of the running Windows executable is unavailable; remove Buda through the stable launcher")
}

func Rollback(executable string) (bool, error) {
	return false, fmt.Errorf("synchronous rollback of the running Windows executable is unavailable; use the stable launcher transaction")
}
