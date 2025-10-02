package session

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestManager_SaveAndLoad(t *testing.T) {
	// Create temporary directory for test
	tmpDir := t.TempDir()

	manager := &Manager{configDir: tmpDir}

	// Create a session
	session := &Session{
		WorkingDir:  "/test/dir",
		CurrentFile: "/test/dir/bookmarks.json",
		MergeHistory: []MergeRecord{
			{
				BaseFile:    "/test/base.json",
				SourceFiles: []string{"/test/firefox.html"},
				Date:        time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Enhanced:    10,
			},
		},
	}

	// Save session
	err := manager.Save(session)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Load session
	loaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded == nil {
		t.Fatal("Load() returned nil session")
	}

	// Verify data
	if loaded.WorkingDir != session.WorkingDir {
		t.Errorf("WorkingDir = %v, want %v", loaded.WorkingDir, session.WorkingDir)
	}

	if loaded.CurrentFile != session.CurrentFile {
		t.Errorf("CurrentFile = %v, want %v", loaded.CurrentFile, session.CurrentFile)
	}

	if len(loaded.MergeHistory) != 1 {
		t.Fatalf("len(MergeHistory) = %v, want 1", len(loaded.MergeHistory))
	}

	record := loaded.MergeHistory[0]
	if record.BaseFile != "/test/base.json" {
		t.Errorf("BaseFile = %v, want /test/base.json", record.BaseFile)
	}

	if record.Enhanced != 10 {
		t.Errorf("Enhanced = %v, want 10", record.Enhanced)
	}
}

func TestManager_Load_NoSession(t *testing.T) {
	tmpDir := t.TempDir()
	manager := &Manager{configDir: tmpDir}

	// Load non-existent session
	session, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v, want nil", err)
	}

	if session != nil {
		t.Error("Load() should return nil for non-existent session")
	}
}

func TestManager_Clear(t *testing.T) {
	tmpDir := t.TempDir()
	manager := &Manager{configDir: tmpDir}

	// Create and save a session
	session := &Session{
		WorkingDir:  "/test/dir",
		CurrentFile: "/test/file.json",
	}

	err := manager.Save(session)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Verify session exists
	if !manager.Exists() {
		t.Fatal("Session should exist after Save()")
	}

	// Clear session
	err = manager.Clear()
	if err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	// Verify session is gone
	if manager.Exists() {
		t.Error("Session should not exist after Clear()")
	}
}

func TestManager_Clear_NoSession(t *testing.T) {
	tmpDir := t.TempDir()
	manager := &Manager{configDir: tmpDir}

	// Clear non-existent session (should not error)
	err := manager.Clear()
	if err != nil {
		t.Errorf("Clear() on non-existent session error = %v, want nil", err)
	}
}

func TestManager_Exists(t *testing.T) {
	tmpDir := t.TempDir()
	manager := &Manager{configDir: tmpDir}

	// Should not exist initially
	if manager.Exists() {
		t.Error("Exists() = true, want false for new manager")
	}

	// Save a session
	session := &Session{WorkingDir: "/test"}
	err := manager.Save(session)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// Should exist now
	if !manager.Exists() {
		t.Error("Exists() = false, want true after Save()")
	}
}

func TestSession_AddMergeRecord(t *testing.T) {
	session := &Session{}

	// Add first merge record
	session.AddMergeRecord("/base1.json", []string{"/source1.html"}, 5)

	if len(session.MergeHistory) != 1 {
		t.Fatalf("len(MergeHistory) = %v, want 1", len(session.MergeHistory))
	}

	record := session.MergeHistory[0]
	if record.BaseFile != "/base1.json" {
		t.Errorf("BaseFile = %v, want /base1.json", record.BaseFile)
	}

	if record.Enhanced != 5 {
		t.Errorf("Enhanced = %v, want 5", record.Enhanced)
	}

	if len(record.SourceFiles) != 1 {
		t.Fatalf("len(SourceFiles) = %v, want 1", len(record.SourceFiles))
	}

	// Add second merge record
	session.AddMergeRecord("/base2.json", []string{"/source2.html", "/source3.html"}, 10)

	if len(session.MergeHistory) != 2 {
		t.Fatalf("len(MergeHistory) = %v, want 2", len(session.MergeHistory))
	}

	record2 := session.MergeHistory[1]
	if record2.Enhanced != 10 {
		t.Errorf("Enhanced = %v, want 10", record2.Enhanced)
	}

	if len(record2.SourceFiles) != 2 {
		t.Fatalf("len(SourceFiles) = %v, want 2", len(record2.SourceFiles))
	}
}

func TestManager_ensureConfigDir(t *testing.T) {
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, "nested", "config", "dir")

	manager := &Manager{configDir: configDir}

	// Ensure directory is created
	err := manager.ensureConfigDir()
	if err != nil {
		t.Fatalf("ensureConfigDir() error = %v", err)
	}

	// Verify directory exists
	info, err := os.Stat(configDir)
	if err != nil {
		t.Fatalf("Config directory not created: %v", err)
	}

	if !info.IsDir() {
		t.Error("Config path is not a directory")
	}
}

func TestManager_LastModified(t *testing.T) {
	tmpDir := t.TempDir()
	manager := &Manager{configDir: tmpDir}

	before := time.Now()

	session := &Session{
		WorkingDir: "/test",
	}

	err := manager.Save(session)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	after := time.Now()

	// Load and check LastModified was set
	loaded, err := manager.Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if loaded.LastModified.IsZero() {
		t.Error("LastModified should be set by Save()")
	}

	if loaded.LastModified.Before(before) || loaded.LastModified.After(after) {
		t.Errorf("LastModified = %v, want between %v and %v", loaded.LastModified, before, after)
	}
}
