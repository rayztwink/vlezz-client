package process

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/rayflow/rayflow-client/apps/backend/internal/logs"
)

func TestStartDoesNotBindProcessToCallerContext(t *testing.T) {
	binaryPath, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable returned error: %v", err)
	}

	manager := NewManager(logs.NewManager(nil, zerolog.Nop()), zerolog.Nop())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if err := manager.Start(ctx, Spec{
		ID:         "helper",
		Name:       "helper",
		BinaryPath: binaryPath,
		Args:       []string{"-test.run=TestHelperProcess", "--", "sleep"},
		Env:        []string{"GO_WANT_HELPER_PROCESS=1"},
	}); err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
	t.Cleanup(func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()
		_ = manager.Stop(stopCtx, "helper")
	})

	time.Sleep(150 * time.Millisecond)
	if status := manager.Status("helper"); status.Status != StatusRunning {
		t.Fatalf("expected process to keep running after caller context cancellation, got %s (%s)", status.Status, status.Error)
	}
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args
	for len(args) > 0 && args[0] != "--" {
		args = args[1:]
	}
	if len(args) < 2 || args[1] != "sleep" {
		os.Exit(2)
	}

	time.Sleep(5 * time.Second)
	os.Exit(0)
}
