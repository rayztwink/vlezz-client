package routing

const (
	ModeDirect = "direct"
	ModeProxy  = "proxy"
	ModeZapret = "zapret"
	ModeHybrid = "hybrid"
	ModeSmart  = "smart"
)

type Rule struct {
	Domain  string
	Mode    string
	Enabled bool
}
