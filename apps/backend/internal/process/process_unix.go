//go:build !windows

package process

import "syscall"

func processSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setpgid: true}
}
