package process

import "time"

type Status string

const (
	StatusStopped Status = "stopped"
	StatusRunning Status = "running"
	StatusFailed  Status = "failed"
)

type Spec struct {
	ID         string
	Name       string
	BinaryPath string
	Args       []string
	WorkingDir string
	Env        []string
}

type Snapshot struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Status    Status    `json:"status"`
	PID       int       `json:"pid,omitempty"`
	StartedAt time.Time `json:"startedAt,omitempty"`
	Error     string    `json:"error,omitempty"`
}
