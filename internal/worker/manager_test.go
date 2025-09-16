package worker

import (
	"testing"
	"time"

	"github.com/entro314-labs/Git-Herd/pkg/types"
)

func TestNew(t *testing.T) {
	config := &types.Config{
		Workers:     5,
		Operation:   types.OperationFetch,
		DryRun:      false,
		Recursive:   true,
		SkipDirty:   true,
		Verbose:     false,
		Timeout:     5 * time.Minute,
		ExcludeDirs: []string{".git", "node_modules"},
	}

	manager := New(config)
	if manager == nil {
		t.Fatal("New returned nil")
	}

	if manager.config != config {
		t.Error("Config not set correctly")
	}

	if manager.logger == nil {
		t.Error("Logger not initialized")
	}

	if manager.scanner == nil {
		t.Error("Scanner not initialized")
	}

	if manager.processor == nil {
		t.Error("Processor not initialized")
	}
}

func TestConfig_OperationType(t *testing.T) {
	tests := []struct {
		name      string
		operation types.OperationType
		expected  types.OperationType
	}{
		{"fetch operation", types.OperationFetch, "fetch"},
		{"pull operation", types.OperationPull, "pull"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.operation != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, tt.operation)
			}
		})
	}
}