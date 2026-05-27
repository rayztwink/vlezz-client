//go:build !windows

package platform

import "os"

func IsAdmin() bool {
	return os.Geteuid() == 0
}
