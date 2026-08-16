//go:build windows

package lifecycle

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func processAliveWindows(pid int) bool {
	command := exec.Command("tasklist", "/FI", "PID eq "+itoa(pid), "/FO", "CSV", "/NH")
	output, err := command.Output()
	if err != nil {
		return false
	}
	needle := `"` + itoa(pid) + `"`
	for _, line := range strings.Split(string(output), "\n") {
		if strings.Contains(line, needle) {
			return true
		}
	}
	return false
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	result := ""
	for value > 0 {
		result = string(rune('0'+value%10)) + result
		value /= 10
	}
	return result
}

func CurrentProcessUser() string { return strings.TrimSpace(os.Getenv("USERNAME")) }

func ProcessExecutable(pid int) (string, error) {
	info, err := processInfoWindows(pid)
	if err != nil {
		return "", err
	}
	return info.Path, nil
}

func TerminateProcess(pid int) error {
	command := exec.Command("taskkill.exe", "/PID", itoa(pid), "/F")
	if output, err := command.CombinedOutput(); err != nil {
		return fmt.Errorf("taskkill failed: %w (%s)", err, strings.TrimSpace(string(output)))
	}
	return nil
}

type processInfo struct {
	Path string `json:"Path"`
	User string `json:"User"`
}

func processInfoWindows(pid int) (processInfo, error) {
	filter := fmt.Sprintf("ProcessId = %d", pid)
	script := fmt.Sprintf("$p=Get-CimInstance Win32_Process -Filter '%s'; if ($null -eq $p) { exit 1 }; $o=Invoke-CimMethod -InputObject $p -MethodName GetOwner; [pscustomobject]@{Path=$p.ExecutablePath;User=$o.User}|ConvertTo-Json -Compress", filter)
	output, err := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script).Output()
	if err != nil {
		if !IsProcessAlive(pid) {
			return processInfo{}, os.ErrProcessDone
		}
		return processInfo{}, err
	}
	var info processInfo
	if err := json.Unmarshal(output, &info); err != nil || strings.TrimSpace(info.Path) == "" {
		return processInfo{}, errors.New("Windows process identity is unavailable")
	}
	if strings.TrimSpace(info.User) == "" {
		info.User = CurrentProcessUser()
	}
	if info.User != CurrentProcessUser() {
		return processInfo{}, errors.New("Windows process belongs to another user")
	}
	return info, nil
}
