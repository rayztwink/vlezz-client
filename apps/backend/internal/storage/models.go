package storage

type AppSettings struct {
	ID                         int    `json:"id"`
	Theme                      string `json:"theme"`
	Language                   string `json:"language"`
	Autostart                  bool   `json:"autostart"`
	ActiveMode                 string `json:"activeMode"`
	DefaultCore                string `json:"defaultCore"`
	LocalProxyPort             int    `json:"localProxyPort"`
	SingBoxPath                string `json:"singBoxPath"`
	XrayPath                   string `json:"xrayPath"`
	ZapretPath                 string `json:"zapretPath"`
	EnableSystemProxyOnConnect bool   `json:"enableSystemProxyOnConnect"`
	PreferredNetworkMode       string `json:"preferredNetworkMode"`
	TUNEnabled                 bool   `json:"tunEnabled"`
	TUNStack                   string `json:"tunStack"`
	TUNAutoRoute               bool   `json:"tunAutoRoute"`
	TUNStrictRoute             bool   `json:"tunStrictRoute"`
	UpdatedAt                  string `json:"updatedAt"`
}

type Node struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Protocol  string `json:"protocol"`
	Address   string `json:"address"`
	Port      int    `json:"port"`
	UUID      string `json:"uuid"`
	Security  string `json:"security"`
	Transport string `json:"transport"`
	RawLink   string `json:"rawLink,omitempty"`
	LatencyMS *int   `json:"latencyMs,omitempty"`
	Country   string `json:"country,omitempty"`
	CreatedAt string `json:"createdAt"`
}

type Subscription struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	URL            string `json:"url"`
	UpdateInterval int    `json:"updateInterval"`
	LastUpdateAt   string `json:"lastUpdateAt,omitempty"`
	CreatedAt      string `json:"createdAt"`
}

type ZapretPreset struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Source      string `json:"source"`
	Command     string `json:"command"`
	Description string `json:"description,omitempty"`
	IsActive    bool   `json:"isActive"`
	UpdatedAt   string `json:"updatedAt"`
}

type RoutingRule struct {
	ID      string `json:"id"`
	Domain  string `json:"domain"`
	Mode    string `json:"mode"`
	Enabled bool   `json:"enabled"`
}

type DiagnosticCheck struct {
	ID        string `json:"id"`
	Target    string `json:"target"`
	Mode      string `json:"mode"`
	Status    string `json:"status"`
	LatencyMS *int   `json:"latencyMs,omitempty"`
	Error     string `json:"error,omitempty"`
	CheckedAt string `json:"checkedAt"`
}

type LogEntry struct {
	ID        string `json:"id"`
	Source    string `json:"source"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	CreatedAt string `json:"createdAt"`
}

type ConnectionState struct {
	ID                int    `json:"id"`
	ActiveMode        string `json:"activeMode"`
	SelectedNodeID    string `json:"selectedNodeId,omitempty"`
	SelectedNodeName  string `json:"selectedNodeName,omitempty"`
	SelectedCore      string `json:"selectedCore"`
	NetworkMode       string `json:"networkMode"`
	LocalProxyAddress string `json:"localProxyAddress"`
	Status            string `json:"status"`
	LastError         string `json:"lastError,omitempty"`
	UpdatedAt         string `json:"updatedAt"`
}

type SystemProxyState struct {
	ID                    int    `json:"id"`
	EnabledByRayFlow      bool   `json:"enabledByRayflow"`
	PreviousProxyEnable   bool   `json:"previousProxyEnable"`
	PreviousProxyServer   string `json:"previousProxyServer"`
	PreviousProxyOverride string `json:"previousProxyOverride"`
	CurrentProxyServer    string `json:"currentProxyServer"`
	UpdatedAt             string `json:"updatedAt"`
}
