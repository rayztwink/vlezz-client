package routing

type SmartDecision struct {
	Mode   string `json:"mode"`
	Reason string `json:"reason"`
}

func DecideSmartRoute(directOK bool, zapretOK bool, proxyOK bool) SmartDecision {
	switch {
	case directOK:
		return SmartDecision{Mode: ModeDirect, Reason: "Direct route is available"}
	case zapretOK:
		return SmartDecision{Mode: ModeZapret, Reason: "Zapret route restored availability"}
	case proxyOK:
		return SmartDecision{Mode: ModeProxy, Reason: "Proxy route is available"}
	default:
		return SmartDecision{Mode: ModeDirect, Reason: "No route passed diagnostics"}
	}
}
