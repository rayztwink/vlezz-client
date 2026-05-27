package core

const (
	NetworkModeLocalProxy  = "local_proxy"
	NetworkModeSystemProxy = "system_proxy"
	NetworkModeTUN         = "tun"
)

type ConnectOptions struct {
	NetworkMode    string
	LocalProxyPort int
	TUNStack       string
	TUNAutoRoute   bool
	TUNStrictRoute bool
}

func (o ConnectOptions) Normalized() ConnectOptions {
	if o.NetworkMode == "" {
		o.NetworkMode = NetworkModeLocalProxy
	}
	if o.LocalProxyPort <= 0 {
		o.LocalProxyPort = 2080
	}
	if o.TUNStack == "" {
		o.TUNStack = "system"
	}
	return o
}
