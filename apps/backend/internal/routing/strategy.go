package routing

func NormalizeMode(mode string) string {
	switch mode {
	case ModeDirect, ModeProxy, ModeZapret, ModeHybrid, ModeSmart:
		return mode
	default:
		return ModeDirect
	}
}
