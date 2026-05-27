//go:build windows

package process

import "syscall"

func processSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true, CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP}
}
