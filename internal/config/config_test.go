package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Beans.IDLength != 4 {
		t.Errorf("IDLength = %d, want 4", cfg.Beans.IDLength)
	}
	if cfg.Beans.Prefix != "" {
		t.Errorf("Prefix = %q, want empty", cfg.Beans.Prefix)
	}
	if cfg.Beans.DefaultStatus != "todo" {
		t.Errorf("DefaultStatus = %q, want \"todo\"", cfg.Beans.DefaultStatus)
	}
	if cfg.Beans.DefaultType != "task" {
		t.Errorf("DefaultType = %q, want \"task\"", cfg.Beans.DefaultType)
	}
	// Both types and statuses are hardcoded
	if len(DefaultTypes) != 5 {
		t.Errorf("len(DefaultTypes) = %d, want 5", len(DefaultTypes))
	}
	if len(DefaultStatuses) != 5 {
		t.Errorf("len(DefaultStatuses) = %d, want 5", len(DefaultStatuses))
	}
}

func TestDefaultWithPrefix(t *testing.T) {
	cfg := DefaultWithPrefix("myapp-")

	if cfg.Beans.Prefix != "myapp-" {
		t.Errorf("Prefix = %q, want \"myapp-\"", cfg.Beans.Prefix)
	}
	// Other defaults should still apply
	if cfg.Beans.IDLength != 4 {
		t.Errorf("IDLength = %d, want 4", cfg.Beans.IDLength)
	}
}

func TestIsValidStatus(t *testing.T) {
	cfg := Default()

	tests := []struct {
		status string
		want   bool
	}{
		{"backlog", true},
		{"todo", true},
		{"in-progress", true},
		{"completed", true},
		{"scrapped", true},
		{"invalid", false},
		{"", false},
		{"TODO", false}, // case sensitive
		// Old status names should no longer be valid
		{"open", false},
		{"done", false},
		{"ready", false},
		{"not-ready", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := cfg.IsValidStatus(tt.status)
			if got != tt.want {
				t.Errorf("IsValidStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestStatusList(t *testing.T) {
	cfg := Default()
	got := cfg.StatusList()
	want := "backlog, todo, in-progress, completed, scrapped"

	if got != want {
		t.Errorf("StatusList() = %q, want %q", got, want)
	}
}

func TestStatusNames(t *testing.T) {
	cfg := Default()
	got := cfg.StatusNames()

	if len(got) != 5 {
		t.Fatalf("len(StatusNames()) = %d, want 5", len(got))
	}
	expected := []string{"backlog", "todo", "in-progress", "completed", "scrapped"}
	for i, name := range expected {
		if got[i] != name {
			t.Errorf("StatusNames()[%d] = %q, want %q", i, got[i], name)
		}
	}
}

func TestGetStatus(t *testing.T) {
	cfg := Default()

	t.Run("existing status", func(t *testing.T) {
		s := cfg.GetStatus("todo")
		if s == nil {
			t.Fatal("GetStatus(\"todo\") = nil, want non-nil")
		}
		if s.Name != "todo" {
			t.Errorf("Name = %q, want \"todo\"", s.Name)
		}
		if s.Color != "green" {
			t.Errorf("Color = %q, want \"green\"", s.Color)
		}
	})

	t.Run("non-existing status", func(t *testing.T) {
		s := cfg.GetStatus("invalid")
		if s != nil {
			t.Errorf("GetStatus(\"invalid\") = %v, want nil", s)
		}
	})

	t.Run("old status names not valid", func(t *testing.T) {
		s := cfg.GetStatus("open")
		if s != nil {
			t.Errorf("GetStatus(\"open\") = %v, want nil (old status name)", s)
		}
		s = cfg.GetStatus("done")
		if s != nil {
			t.Errorf("GetStatus(\"done\") = %v, want nil (old status name)", s)
		}
		s = cfg.GetStatus("ready")
		if s != nil {
			t.Errorf("GetStatus(\"ready\") = %v, want nil (old status name)", s)
		}
	})
}

func TestGetDefaultStatus(t *testing.T) {
	cfg := Default()
	got := cfg.GetDefaultStatus()

	if got != "todo" {
		t.Errorf("GetDefaultStatus() = %q, want \"todo\"", got)
	}
}

func TestGetDefaultType(t *testing.T) {
	cfg := Default()
	got := cfg.GetDefaultType()

	if got != "task" {
		t.Errorf("GetDefaultType() = %q, want \"task\"", got)
	}
}

func TestIsArchiveStatus(t *testing.T) {
	cfg := Default()

	tests := []struct {
		status string
		want   bool
	}{
		{"completed", true},
		{"scrapped", true},
		{"backlog", false},
		{"todo", false},
		{"in-progress", false},
		{"invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.status, func(t *testing.T) {
			got := cfg.IsArchiveStatus(tt.status)
			if got != tt.want {
				t.Errorf("IsArchiveStatus(%q) = %v, want %v", tt.status, got, tt.want)
			}
		})
	}
}

func TestLoadNonExistent(t *testing.T) {
	// Load from non-existent directory should return defaults
	cfg, err := Load("/nonexistent/path/that/does/not/exist")
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	// Should have default values
	if cfg.Beans.IDLength != 4 {
		t.Errorf("IDLength = %d, want 4", cfg.Beans.IDLength)
	}
}

func TestLoadAndSave(t *testing.T) {
	// Create temp directory
	tmpDir := t.TempDir()

	// Create a config (statuses are no longer stored in config)
	cfg := &Config{
		Beans: BeansConfig{
			Prefix:      "test-",
			IDLength:    6,
			DefaultType: "bug",
		},
	}

	// Save it
	if err := cfg.Save(tmpDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify file exists
	configPath := filepath.Join(tmpDir, ConfigFile)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("config file was not created")
	}

	// Load it back
	loaded, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify values
	if loaded.Beans.Prefix != "test-" {
		t.Errorf("Prefix = %q, want \"test-\"", loaded.Beans.Prefix)
	}
	if loaded.Beans.IDLength != 6 {
		t.Errorf("IDLength = %d, want 6", loaded.Beans.IDLength)
	}
	if loaded.Beans.DefaultType != "bug" {
		t.Errorf("DefaultType = %q, want \"bug\"", loaded.Beans.DefaultType)
	}
	// Statuses are hardcoded, not stored in config
	if len(loaded.StatusNames()) != 5 {
		t.Errorf("len(StatusNames()) = %d, want 5", len(loaded.StatusNames()))
	}
}

func TestLoadAppliesDefaults(t *testing.T) {
	// Create temp directory with minimal config
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, ConfigFile)

	// Write minimal config (missing id_length and default_type)
	minimalConfig := `beans:
  prefix: "my-"
`
	if err := os.WriteFile(configPath, []byte(minimalConfig), 0644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	// Load it
	cfg, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Verify defaults were applied
	if cfg.Beans.IDLength != 4 {
		t.Errorf("IDLength default not applied: got %d, want 4", cfg.Beans.IDLength)
	}
	// Statuses are hardcoded, always 5
	if len(cfg.StatusNames()) != 5 {
		t.Errorf("Hardcoded statuses: got %d, want 5", len(cfg.StatusNames()))
	}
	// DefaultStatus is always "todo"
	if cfg.GetDefaultStatus() != "todo" {
		t.Errorf("DefaultStatus: got %q, want \"todo\"", cfg.GetDefaultStatus())
	}
	// DefaultType should be first type name when not specified
	if cfg.Beans.DefaultType != "milestone" {
		t.Errorf("DefaultType default not applied: got %q, want \"milestone\"", cfg.Beans.DefaultType)
	}
}

func TestStatusesAreHardcoded(t *testing.T) {
	// Statuses are hardcoded and not configurable (like types)
	// Verify that any config only uses hardcoded statuses
	cfg := Default()

	// All hardcoded statuses should be valid
	hardcodedStatuses := []string{"backlog", "todo", "in-progress", "completed", "scrapped"}
	for _, status := range hardcodedStatuses {
		if !cfg.IsValidStatus(status) {
			t.Errorf("IsValidStatus(%q) = false, want true", status)
		}
	}

	// Archive statuses should be completed and scrapped
	if !cfg.IsArchiveStatus("completed") {
		t.Error("IsArchiveStatus(\"completed\") = false, want true")
	}
	if !cfg.IsArchiveStatus("scrapped") {
		t.Error("IsArchiveStatus(\"scrapped\") = false, want true")
	}
	if cfg.IsArchiveStatus("todo") {
		t.Error("IsArchiveStatus(\"todo\") = true, want false")
	}
}

func TestIsValidType(t *testing.T) {
	cfg := Default()

	tests := []struct {
		typeName string
		want     bool
	}{
		{"epic", true},
		{"milestone", true},
		{"feature", true},
		{"bug", true},
		{"task", true},
		{"invalid", false},
		{"", false},
		{"TASK", false}, // case sensitive
	}

	for _, tt := range tests {
		t.Run(tt.typeName, func(t *testing.T) {
			got := cfg.IsValidType(tt.typeName)
			if got != tt.want {
				t.Errorf("IsValidType(%q) = %v, want %v", tt.typeName, got, tt.want)
			}
		})
	}
}

func TestTypeList(t *testing.T) {
	cfg := Default()
	got := cfg.TypeList()
	want := "milestone, epic, bug, feature, task"

	if got != want {
		t.Errorf("TypeList() = %q, want %q", got, want)
	}
}

func TestGetType(t *testing.T) {
	cfg := Default()

	t.Run("existing type", func(t *testing.T) {
		typ := cfg.GetType("bug")
		if typ == nil {
			t.Fatal("GetType(\"bug\") = nil, want non-nil")
		}
		if typ.Name != "bug" {
			t.Errorf("Name = %q, want \"bug\"", typ.Name)
		}
		if typ.Color != "red" {
			t.Errorf("Color = %q, want \"red\"", typ.Color)
		}
	})

	t.Run("non-existing type", func(t *testing.T) {
		// GetType returns nil for unknown types
		typ := cfg.GetType("invalid-type")
		if typ != nil {
			t.Errorf("GetType(\"invalid-type\") = %v, want nil", typ)
		}
	})

	t.Run("all hardcoded types exist", func(t *testing.T) {
		expectedTypes := []string{"milestone", "epic", "bug", "feature", "task"}
		for _, typeName := range expectedTypes {
			typ := cfg.GetType(typeName)
			if typ == nil {
				t.Errorf("GetType(%q) = nil, want non-nil", typeName)
			}
		}
	})
}

func TestTypesAreHardcoded(t *testing.T) {
	// Types are hardcoded and not stored in config
	// Verify that saving and loading a config doesn't affect types

	tmpDir := t.TempDir()

	cfg := &Config{
		Beans: BeansConfig{
			Prefix:      "test-",
			IDLength:    4,
			DefaultType: "task",
		},
	}

	// Save it
	if err := cfg.Save(tmpDir); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load it back
	loaded, err := Load(tmpDir)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	// Types should always come from DefaultTypes, not config
	if len(loaded.TypeNames()) != 5 {
		t.Errorf("len(TypeNames()) = %d, want 5", len(loaded.TypeNames()))
	}

	// All default types should be accessible
	for _, typeName := range []string{"milestone", "epic", "bug", "feature", "task"} {
		if !loaded.IsValidType(typeName) {
			t.Errorf("IsValidType(%q) = false, want true", typeName)
		}
	}

	// Statuses should also be hardcoded
	if len(loaded.StatusNames()) != 5 {
		t.Errorf("len(StatusNames()) = %d, want 5", len(loaded.StatusNames()))
	}
}

func TestTypeDescriptions(t *testing.T) {
	t.Run("hardcoded types have descriptions", func(t *testing.T) {
		cfg := Default()

		expectedDescriptions := map[string]string{
			"epic":      "A thematic container for related work; should have child beans, not be worked on directly",
			"milestone": "A target release or checkpoint; group work that should ship together",
			"feature":   "A user-facing capability or enhancement",
			"bug":       "Something that is broken and needs fixing",
			"task":      "A concrete piece of work to complete (eg. a chore, or a sub-task for a feature)",
		}

		for typeName, expectedDesc := range expectedDescriptions {
			typ := cfg.GetType(typeName)
			if typ == nil {
				t.Errorf("GetType(%q) = nil, want non-nil", typeName)
				continue
			}
			if typ.Description != expectedDesc {
				t.Errorf("Type %q description = %q, want %q", typeName, typ.Description, expectedDesc)
			}
		}
	})

	t.Run("types in config file are ignored", func(t *testing.T) {
		// Even if a config file has custom types, they should be ignored
		// and hardcoded types should be used instead
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ConfigFile)

		// Config with custom types (should be ignored)
		configYAML := `beans:
  prefix: "test-"
  id_length: 4
  default_status: open
statuses:
  - name: open
    color: green
types:
  - name: custom-type
    color: pink
    description: "This should be ignored"
`
		if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
			t.Fatalf("WriteFile error = %v", err)
		}

		loaded, err := Load(tmpDir)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Custom type should not be valid
		if loaded.IsValidType("custom-type") {
			t.Error("IsValidType(\"custom-type\") = true, want false (custom types should be ignored)")
		}

		// Hardcoded types should still work
		if !loaded.IsValidType("bug") {
			t.Error("IsValidType(\"bug\") = false, want true")
		}
	})
}

func TestStatusDescriptions(t *testing.T) {
	t.Run("hardcoded statuses have descriptions", func(t *testing.T) {
		cfg := Default()

		expectedDescriptions := map[string]string{
			"backlog":     "Not yet ready to be worked on",
			"todo":        "Ready to be worked on",
			"in-progress": "Currently being worked on",
			"completed":   "Finished successfully",
			"scrapped":    "Will not be done",
		}

		for statusName, expectedDesc := range expectedDescriptions {
			status := cfg.GetStatus(statusName)
			if status == nil {
				t.Errorf("GetStatus(%q) = nil, want non-nil", statusName)
				continue
			}
			if status.Description != expectedDesc {
				t.Errorf("Status %q description = %q, want %q", statusName, status.Description, expectedDesc)
			}
		}
	})

	t.Run("statuses in config file are ignored", func(t *testing.T) {
		// Even if a config file has custom statuses, they should be ignored
		// and hardcoded statuses should be used instead
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, ConfigFile)

		// Config with custom statuses (should be ignored)
		configYAML := `beans:
  prefix: "test-"
  id_length: 4
statuses:
  - name: custom-status
    color: pink
    description: "This should be ignored"
`
		if err := os.WriteFile(configPath, []byte(configYAML), 0644); err != nil {
			t.Fatalf("WriteFile error = %v", err)
		}

		loaded, err := Load(tmpDir)
		if err != nil {
			t.Fatalf("Load() error = %v", err)
		}

		// Custom status should not be valid
		if loaded.IsValidStatus("custom-status") {
			t.Error("IsValidStatus(\"custom-status\") = true, want false (custom statuses should be ignored)")
		}

		// Hardcoded statuses should still work
		if !loaded.IsValidStatus("todo") {
			t.Error("IsValidStatus(\"todo\") = false, want true")
		}
	})
}
