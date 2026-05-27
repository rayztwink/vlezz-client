//go:build windows

package systemproxy

import (
	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const internetSettingsPath = `Software\Microsoft\Windows\CurrentVersion\Internet Settings`

func platformSupported() bool {
	return true
}

func readPlatformProxy() (PlatformState, error) {
	key, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsPath, registry.QUERY_VALUE)
	if err != nil {
		return PlatformState{}, err
	}
	defer key.Close()
	enable, _, err := key.GetIntegerValue("ProxyEnable")
	if err != nil && err != registry.ErrNotExist {
		return PlatformState{}, err
	}
	server, _, err := key.GetStringValue("ProxyServer")
	if err != nil && err != registry.ErrNotExist {
		return PlatformState{}, err
	}
	override, _, err := key.GetStringValue("ProxyOverride")
	if err != nil && err != registry.ErrNotExist {
		return PlatformState{}, err
	}
	return PlatformState{ProxyEnable: enable == 1, ProxyServer: server, ProxyOverride: override}, nil
}

func writePlatformProxy(state PlatformState) error {
	key, err := registry.OpenKey(registry.CURRENT_USER, internetSettingsPath, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	enable := uint32(0)
	if state.ProxyEnable {
		enable = 1
	}
	if err := key.SetDWordValue("ProxyEnable", enable); err != nil {
		return err
	}
	if err := key.SetStringValue("ProxyServer", state.ProxyServer); err != nil {
		return err
	}
	if err := key.SetStringValue("ProxyOverride", state.ProxyOverride); err != nil {
		return err
	}
	notifyProxyChange()
	return nil
}

func notifyProxyChange() {
	wininet := windows.NewLazySystemDLL("wininet.dll")
	internetSetOption := wininet.NewProc("InternetSetOptionW")
	const internetOptionSettingsChanged = 39
	const internetOptionRefresh = 37
	internetSetOption.Call(0, internetOptionSettingsChanged, 0, 0)
	internetSetOption.Call(0, internetOptionRefresh, 0, 0)
}
