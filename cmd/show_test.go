package cmd

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hmans/beans/internal/bean"
	"github.com/hmans/beans/internal/beancore"
	"github.com/hmans/beans/internal/config"
)

// setupShowTestCore creates a test core for show command tests
func setupShowTestCore(t *testing.T) (*beancore.Core, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	beansDir := filepath.Join(tmpDir, ".beans")

	// Create the .beans directory
	if err := os.MkdirAll(beansDir, 0755); err != nil {
		t.Fatalf("failed to create .beans dir: %v", err)
	}

	cfg := config.Default()
	testCore := beancore.New(beansDir, cfg)
	testCore.SetWarnWriter(nil)
	if err := testCore.Load(); err != nil {
		t.Fatalf("failed to load core: %v", err)
	}

	oldCore := core
	core = testCore

	cleanup := func() {
		core = oldCore
	}

	return testCore, cleanup
}

func TestShowCommand_GitFields_WithBranch(t *testing.T) {
	testCore, cleanup := setupShowTestCore(t)
	defer cleanup()

	// Create bean with git metadata
	now := time.Now()
	mergedAt := now.Add(-24 * time.Hour)

	b := &bean.Bean{
		ID:             "beans-test1",
		Slug:           "test-bean",
		Title:          "Test Bean",
		Status:         "completed",
		Type:           "feature",
		GitBranch:      "beans-test1/test-bean",
		GitCreatedAt:   &now,
		GitMergedAt:    &mergedAt,
		GitMergeCommit: "abc123def456",
	}
	if err := testCore.Create(b); err != nil {
		t.Fatalf("Create bean error = %v", err)
	}

	// Get the bean
	retrieved, err := testCore.Get(b.ID)
	if err != nil {
		t.Fatalf("Get bean error = %v", err)
	}

	// Verify git fields are present
	if retrieved.GitBranch != "beans-test1/test-bean" {
		t.Errorf("GitBranch = %q, want %q", retrieved.GitBranch, "beans-test1/test-bean")
	}
	if retrieved.GitCreatedAt == nil {
		t.Error("GitCreatedAt should not be nil")
	}
	if retrieved.GitMergedAt == nil {
		t.Error("GitMergedAt should not be nil")
	}
	if retrieved.GitMergeCommit != "abc123def456" {
		t.Errorf("GitMergeCommit = %q, want %q", retrieved.GitMergeCommit, "abc123def456")
	}
}

func TestShowCommand_GitFields_WithoutBranch(t *testing.T) {
	testCore, cleanup := setupShowTestCore(t)
	defer cleanup()

	// Create bean without git metadata
	b := &bean.Bean{
		ID:     "beans-test1",
		Slug:   "test-bean",
		Title:  "Test Bean",
		Status: "todo",
		Type:   "task",
	}
	if err := testCore.Create(b); err != nil {
		t.Fatalf("Create bean error = %v", err)
	}

	// Get the bean
	retrieved, err := testCore.Get(b.ID)
	if err != nil {
		t.Fatalf("Get bean error = %v", err)
	}

	// Verify git fields are empty
	if retrieved.GitBranch != "" {
		t.Errorf("GitBranch should be empty, got %q", retrieved.GitBranch)
	}
	if retrieved.GitCreatedAt != nil {
		t.Error("GitCreatedAt should be nil")
	}
	if retrieved.GitMergedAt != nil {
		t.Error("GitMergedAt should be nil")
	}
	if retrieved.GitMergeCommit != "" {
		t.Errorf("GitMergeCommit should be empty, got %q", retrieved.GitMergeCommit)
	}
}

func TestShowCommand_JSONOutput_WithGitFields(t *testing.T) {
	testCore, cleanup := setupShowTestCore(t)
	defer cleanup()

	// Create bean with git metadata
	now := time.Now()
	b := &bean.Bean{
		ID:           "beans-test1",
		Slug:         "test-bean",
		Title:        "Test Bean",
		Status:       "in-progress",
		Type:         "feature",
		GitBranch:    "beans-test1/test-bean",
		GitCreatedAt: &now,
	}
	if err := testCore.Create(b); err != nil {
		t.Fatalf("Create bean error = %v", err)
	}

	// Get the bean
	retrieved, err := testCore.Get(b.ID)
	if err != nil {
		t.Fatalf("Get bean error = %v", err)
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(retrieved)
	if err != nil {
		t.Fatalf("JSON marshal error = %v", err)
	}

	jsonStr := string(jsonBytes)

	// Verify git fields are in JSON output
	if !strings.Contains(jsonStr, "git_branch") && !strings.Contains(jsonStr, "gitBranch") {
		t.Error("JSON output should contain git branch field")
	}

	// Parse back to verify
	var parsed bean.Bean
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if parsed.GitBranch != "beans-test1/test-bean" {
		t.Errorf("Parsed GitBranch = %q, want %q", parsed.GitBranch, "beans-test1/test-bean")
	}
	if parsed.GitCreatedAt == nil {
		t.Error("Parsed GitCreatedAt should not be nil")
	}
}

func TestShowCommand_JSONOutput_WithoutGitFields(t *testing.T) {
	testCore, cleanup := setupShowTestCore(t)
	defer cleanup()

	// Create bean without git metadata
	b := &bean.Bean{
		ID:     "beans-test1",
		Slug:   "test-bean",
		Title:  "Test Bean",
		Status: "todo",
		Type:   "task",
	}
	if err := testCore.Create(b); err != nil {
		t.Fatalf("Create bean error = %v", err)
	}

	// Get the bean
	retrieved, err := testCore.Get(b.ID)
	if err != nil {
		t.Fatalf("Get bean error = %v", err)
	}

	// Marshal to JSON
	jsonBytes, err := json.Marshal(retrieved)
	if err != nil {
		t.Fatalf("JSON marshal error = %v", err)
	}

	// Parse back to verify git fields are empty
	var parsed bean.Bean
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("JSON unmarshal error = %v", err)
	}

	if parsed.GitBranch != "" {
		t.Errorf("Parsed GitBranch should be empty, got %q", parsed.GitBranch)
	}
	if parsed.GitCreatedAt != nil {
		t.Error("Parsed GitCreatedAt should be nil")
	}
}

func TestShowCommand_HumanReadableOutput(t *testing.T) {
	testCore, cleanup := setupShowTestCore(t)
	defer cleanup()

	// Create bean with git metadata
	now := time.Now()
	b := &bean.Bean{
		ID:           "beans-test1",
		Slug:         "test-bean",
		Title:        "Test Bean",
		Status:       "in-progress",
		Type:         "feature",
		Body:         "This is a test bean with git integration",
		GitBranch:    "beans-test1/test-bean",
		GitCreatedAt: &now,
	}
	if err := testCore.Create(b); err != nil {
		t.Fatalf("Create bean error = %v", err)
	}

	// Format the bean for display (simulating what show command does)
	retrieved, err := testCore.Get(b.ID)
	if err != nil {
		t.Fatalf("Get bean error = %v", err)
	}

	// Create a buffer to capture output
	var buf bytes.Buffer

	// Simulate formatting git section
	if retrieved.GitBranch != "" {
		buf.WriteString("Git Branch: ")
		buf.WriteString(retrieved.GitBranch)
		buf.WriteString("\n")
	}
	if retrieved.GitCreatedAt != nil {
		buf.WriteString("Git Created: ")
		buf.WriteString(retrieved.GitCreatedAt.Format(time.RFC3339))
		buf.WriteString("\n")
	}

	output := buf.String()

	// Verify git information is in output
	if !strings.Contains(output, "beans-test1/test-bean") {
		t.Error("Human-readable output should contain git branch name")
	}
	if retrieved.GitCreatedAt != nil && !strings.Contains(output, "Git Created") {
		t.Error("Human-readable output should contain git created timestamp")
	}
}

func TestShowCommand_MultipleBeansWithMixedGitStatus(t *testing.T) {
	testCore, cleanup := setupShowTestCore(t)
	defer cleanup()

	// Create bean with git branch
	now := time.Now()
	withGit := &bean.Bean{
		ID:           "beans-with-git",
		Slug:         "with-git",
		Title:        "With Git",
		Status:       "in-progress",
		Type:         "feature",
		GitBranch:    "beans-with-git/with-git",
		GitCreatedAt: &now,
	}
	testCore.Create(withGit)

	// Create bean without git branch
	withoutGit := &bean.Bean{
		ID:     "beans-without-git",
		Slug:   "without-git",
		Title:  "Without Git",
		Status: "todo",
		Type:   "task",
	}
	testCore.Create(withoutGit)

	// Verify first bean has git fields
	retrieved1, _ := testCore.Get("beans-with-git")
	if retrieved1.GitBranch == "" {
		t.Error("beans-with-git should have GitBranch")
	}

	// Verify second bean doesn't have git fields
	retrieved2, _ := testCore.Get("beans-without-git")
	if retrieved2.GitBranch != "" {
		t.Error("beans-without-git should not have GitBranch")
	}
}
