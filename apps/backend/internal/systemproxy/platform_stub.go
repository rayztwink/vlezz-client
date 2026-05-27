//go:build !windows

package systemproxy

import (
	"bytes"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

func platformSupported() bool {
	// macOS and Linux are supported
	return runtime.GOOS == "darwin" || runtime.GOOS == "linux"
}

func readPlatformProxy() (PlatformState, error) {
	if runtime.GOOS == "darwin" {
		return readMacProxy()
	}
	if runtime.GOOS == "linux" {
		return readLinuxProxy()
	}
	return PlatformState{}, fmt.Errorf("system proxy is not supported on %s", runtime.GOOS)
}

func writePlatformProxy(state PlatformState) error {
	if runtime.GOOS == "darwin" {
		return writeMacProxy(state)
	}
	if runtime.GOOS == "linux" {
		return writeLinuxProxy(state)
	}
	return fmt.Errorf("system proxy is not supported on %s", runtime.GOOS)
}

// --- macOS (networksetup) ---

func readMacProxy() (PlatformState, error) {
	service := "Wi-Fi" // Default service
	cmd := exec.Command("networksetup", "-getsocksfirewallproxy", service)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return PlatformState{}, err
	}
	output := out.String()
	state := PlatformState{}
	for _, line := range strings.Split(output, "\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "Enabled":
			state.ProxyEnable = (strings.ToLower(val) == "yes")
		case "Server":
			state.ProxyServer = val
		case "Port":
			state.ProxyServer = netJoin(state.ProxyServer, val)
		}
	}
	return state, nil
}

func writeMacProxy(state PlatformState) error {
	service := "Wi-Fi" // Default service
	if state.ProxyEnable {
		host, portStr, err := netSplit(state.ProxyServer)
		if err != nil {
			host = "127.0.0.1"
			portStr = "2080"
		}
		cmd1 := exec.Command("networksetup", "-setsocksfirewallproxy", service, host, portStr)
		if err := cmd1.Run(); err != nil {
			return err
		}
		cmd2 := exec.Command("networksetup", "-setsocksfirewallproxystate", service, "on")
		return cmd2.Run()
	} else {
		cmd := exec.Command("networksetup", "-setsocksfirewallproxystate", service, "off")
		return cmd.Run()
	}
}

// --- Linux (GNOME gsettings) ---

func readLinuxProxy() (PlatformState, error) {
	if _, err := exec.LookPath("gsettings"); err != nil {
		return PlatformState{}, fmt.Errorf("gsettings is not installed")
	}

	modeCmd := exec.Command("gsettings", "get", "org.gnome.system.proxy", "mode")
	modeBytes, err := modeCmd.Output()
	if err != nil {
		return PlatformState{}, err
	}
	mode := strings.Trim(strings.TrimSpace(string(modeBytes)), "'\"")

	state := PlatformState{
		ProxyEnable: (mode == "manual"),
	}

	hostCmd := exec.Command("gsettings", "get", "org.gnome.system.proxy.socks", "host")
	hostBytes, _ := hostCmd.Output()
	host := strings.Trim(strings.TrimSpace(string(hostBytes)), "'\"")

	portCmd := exec.Command("gsettings", "get", "org.gnome.system.proxy.socks", "port")
	portBytes, _ := portCmd.Output()
	port := strings.TrimSpace(string(portBytes))

	state.ProxyServer = netJoin(host, port)
	return state, nil
}

func writeLinuxProxy(state PlatformState) error {
	if _, err := exec.LookPath("gsettings"); err != nil {
		return fmt.Errorf("gsettings is not installed")
	}

	if state.ProxyEnable {
		host, portStr, err := netSplit(state.ProxyServer)
		if err != nil {
			host = "127.0.0.1"
			portStr = "2080"
		}
		if err := exec.Command("gsettings", "set", "org.gnome.system.proxy.socks", "host", host).Run(); err != nil {
			return err
		}
		if err := exec.Command("gsettings", "set", "org.gnome.system.proxy.socks", "port", portStr).Run(); err != nil {
			return err
		}
		return exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "manual").Run()
	} else {
		return exec.Command("gsettings", "set", "org.gnome.system.proxy", "mode", "none").Run()
	}
}

// --- Helpers ---

func netJoin(host, port string) string {
	if host == "" {
		return ""
	}
	if port == "" || port == "0" {
		return host
	}
	return host + ":" + port
}

func netSplit(addr string) (string, string, error) {
	parts := strings.SplitN(addr, ":", 2)
	if len(parts) == 2 {
		return parts[0], parts[1], nil
	}
	if len(parts) == 1 {
		return parts[0], "", nil
	}
	return "", "", fmt.Errorf("empty address")
}
