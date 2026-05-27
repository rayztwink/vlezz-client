package process

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/rayflow/rayflow-client/apps/backend/internal/logs"
)

type managedProcess struct {
	spec      Spec
	cmd       *exec.Cmd
	cancel    context.CancelFunc
	done      chan error
	status    Status
	startedAt time.Time
	lastError string
}

type Manager struct {
	mu        sync.Mutex
	processes map[string]*managedProcess
	logs      *logs.Manager
	logger    zerolog.Logger
}

func NewManager(logManager *logs.Manager, logger zerolog.Logger) *Manager {
	return &Manager{
		processes: make(map[string]*managedProcess),
		logs:      logManager,
		logger:    logger,
	}
}

func (m *Manager) Start(ctx context.Context, spec Spec) error {
	if spec.ID == "" {
		return errors.New("process id is required")
	}
	if err := ValidateBinaryPath(spec.BinaryPath); err != nil {
		return err
	}

	m.mu.Lock()
	if existing, ok := m.processes[spec.ID]; ok && existing.status == StatusRunning {
		m.mu.Unlock()
		return fmt.Errorf("process %q is already running", spec.ID)
	}
	m.mu.Unlock()

	processCtx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(processCtx, spec.BinaryPath, spec.Args...)
	cmd.SysProcAttr = processSysProcAttr()
	if spec.WorkingDir != "" {
		cmd.Dir = spec.WorkingDir
	}
	if len(spec.Env) > 0 {
		cmd.Env = append(os.Environ(), spec.Env...)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		cancel()
		return err
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return err
	}

	mp := &managedProcess{
		spec:      spec,
		cmd:       cmd,
		cancel:    cancel,
		done:      make(chan error, 1),
		status:    StatusRunning,
		startedAt: time.Now().UTC(),
	}

	m.mu.Lock()
	m.processes[spec.ID] = mp
	m.mu.Unlock()

	m.logs.Info(spec.ID, fmt.Sprintf("%s started with pid %d", spec.Name, cmd.Process.Pid))
	go m.scanOutput(spec.ID, "info", stdout)
	go m.scanOutput(spec.ID, "error", stderr)
	go m.wait(spec.ID, cmd, mp.done)
	return nil
}

func (m *Manager) Stop(ctx context.Context, id string) error {
	m.mu.Lock()
	mp, ok := m.processes[id]
	if !ok || mp.status != StatusRunning {
		m.mu.Unlock()
		return nil
	}
	cmd := mp.cmd
	cancel := mp.cancel
	m.mu.Unlock()

	cancel()
	done := mp.done

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(250 * time.Millisecond):
		_ = terminateProcess(cmd)
	case <-done:
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	case <-done:
	}

	m.mu.Lock()
	if current, ok := m.processes[id]; ok {
		current.status = StatusStopped
		current.lastError = ""
	}
	m.mu.Unlock()

	m.logs.Info(id, "process stopped")
	return nil
}

func (m *Manager) Restart(ctx context.Context, spec Spec) error {
	if err := m.Stop(ctx, spec.ID); err != nil {
		return err
	}
	return m.Start(ctx, spec)
}

func (m *Manager) StopAll(ctx context.Context) error {
	m.mu.Lock()
	ids := make([]string, 0, len(m.processes))
	for id, p := range m.processes {
		if p.status == StatusRunning {
			ids = append(ids, id)
		}
	}
	m.mu.Unlock()

	for _, id := range ids {
		if err := m.Stop(ctx, id); err != nil {
			return err
		}
	}
	return nil
}

func (m *Manager) Status(id string) Snapshot {
	m.mu.Lock()
	defer m.mu.Unlock()
	mp, ok := m.processes[id]
	if !ok {
		return Snapshot{ID: id, Status: StatusStopped}
	}
	pid := 0
	if mp.cmd != nil && mp.cmd.Process != nil {
		pid = mp.cmd.Process.Pid
	}
	return Snapshot{
		ID:        id,
		Name:      mp.spec.Name,
		Status:    mp.status,
		PID:       pid,
		StartedAt: mp.startedAt,
		Error:     mp.lastError,
	}
}

func (m *Manager) scanOutput(source string, level string, reader any) {
	r, ok := reader.(interface {
		Read([]byte) (int, error)
	})
	if !ok {
		return
	}
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		m.logs.Record(context.Background(), source, level, scanner.Text())
	}
}

func (m *Manager) wait(id string, cmd *exec.Cmd, done chan error) {
	err := cmd.Wait()
	done <- err
	close(done)
	m.mu.Lock()
	defer m.mu.Unlock()
	mp, ok := m.processes[id]
	if !ok {
		return
	}
	if err != nil {
		mp.status = StatusFailed
		mp.lastError = err.Error()
		m.logs.Error(id, fmt.Sprintf("process exited with error: %v", err))
		return
	}
	mp.status = StatusStopped
	mp.lastError = ""
	m.logs.Info(id, "process exited")
}

func ValidateBinaryPath(path string) error {
	if path == "" {
		return errors.New("binary path is required")
	}
	if !filepath.IsAbs(path) {
		return fmt.Errorf("binary path must be absolute: %s", path)
	}
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("binary path is not accessible: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("binary path points to a directory: %s", path)
	}

	// Security Hardening: Validate that the binary is in the whitelist of allowed cores
	base := filepath.Base(path)
	allowed := false
	whitelist := []string{"sing-box", "sing-box.exe", "xray", "xray.exe", "zapret", "zapret.exe", "pkexec", "osascript"}
	for _, name := range whitelist {
		if strings.EqualFold(base, name) {
			allowed = true
			break
		}
	}
	// Allow Go test binaries to run during Go testing
	if !allowed && (strings.HasSuffix(strings.ToLower(base), ".test.exe") || strings.HasSuffix(strings.ToLower(base), ".test")) {
		allowed = true
	}
	if !allowed {
		return fmt.Errorf("security violation: binary file name %q is not in the allowed whitelist (sing-box, xray, zapret, pkexec, osascript)", base)
	}

	return nil
}

func terminateProcess(cmd *exec.Cmd) error {
	if cmd == nil || cmd.Process == nil {
		return nil
	}
	return cmd.Process.Signal(syscall.SIGTERM)
}
