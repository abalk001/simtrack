package simtrack

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// initTestRepo creates a throwaway Git repo with one commit and returns its path.
func initTestRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	mustRun(t, dir, "git", "init", "-q")
	mustRun(t, dir, "git", "config", "user.email", "t@t")
	mustRun(t, dir, "git", "config", "user.name", "t")
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	mustRun(t, dir, "git", "add", ".")
	mustRun(t, dir, "git", "commit", "-qm", "init")
	return dir
}

func mustRun(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%s %v: %v\n%s", name, args, err, out)
	}
}

func latestManifest(t *testing.T, projectRoot string) *Manifest {
	t.Helper()
	runsDir := filepath.Join(projectRoot, "runs")
	entries, err := os.ReadDir(runsDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) == 0 {
		t.Fatal("no runs found")
	}
	latest := entries[0].Name()
	for _, e := range entries {
		if e.Name() > latest {
			latest = e.Name()
		}
	}
	m, err := readManifest(filepath.Join(runsDir, latest, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// Test 1 — success.
func TestRun_Success(t *testing.T) {
	dir := initTestRepo(t)
	code, err := Run(RunOptions{Cwd: dir, Args: []string{"echo", "hello"}})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("exit code = %d, want 0", code)
	}

	m := latestManifest(t, dir)
	if m.Status != "completed" {
		t.Errorf("status = %q, want completed", m.Status)
	}
	if m.GitDirty {
		t.Error("git_dirty = true, want false")
	}
	if m.ExitCode == nil || *m.ExitCode != 0 {
		t.Errorf("exit_code = %v, want 0", m.ExitCode)
	}

	stdout, err := os.ReadFile(filepath.Join(dir, "runs", m.RunID, "stdout.log"))
	if err != nil {
		t.Fatal(err)
	}
	if string(stdout) != "hello\n" {
		t.Errorf("stdout = %q, want %q", stdout, "hello\n")
	}
}

// Test 2 — failure.
func TestRun_Failure(t *testing.T) {
	dir := initTestRepo(t)
	code, err := Run(RunOptions{Cwd: dir, Args: []string{"false"}})
	if err != nil {
		t.Fatal(err)
	}
	if code == 0 {
		t.Fatal("expected non-zero exit code")
	}

	m := latestManifest(t, dir)
	if m.Status != "failed" {
		t.Errorf("status = %q, want failed", m.Status)
	}
	if m.ExitCode == nil || *m.ExitCode != 1 {
		t.Errorf("exit_code = %v, want 1", m.ExitCode)
	}
}

// Test 3 — dirty working tree.
func TestRun_Dirty(t *testing.T) {
	dir := initTestRepo(t)
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("modified\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Run(RunOptions{Cwd: dir, Args: []string{"true"}}); err != nil {
		t.Fatal(err)
	}
	m := latestManifest(t, dir)
	if !m.GitDirty {
		t.Error("git_dirty = false, want true")
	}
}

// Test 4 — command with arguments.
func TestRun_WithArgs(t *testing.T) {
	dir := initTestRepo(t)
	code, err := Run(RunOptions{Cwd: dir, Args: []string{"sh", "-c", "echo $((2+2))"}})
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 {
		t.Fatalf("exit code = %d", code)
	}
	m := latestManifest(t, dir)
	if m.Command != "sh -c echo $((2+2))" {
		t.Errorf("command = %q", m.Command)
	}
}

// Test 5 — outside a Git repo.
func TestRun_NotInGitRepo(t *testing.T) {
	dir := t.TempDir()
	_, err := Run(RunOptions{Cwd: dir, Args: []string{"true"}})
	if err == nil {
		t.Fatal("expected error outside git repo")
	}
}

// Test 6 — manifest schema is exactly Phase 0's.
func TestManifest_Schema(t *testing.T) {
	dir := initTestRepo(t)
	if _, err := Run(RunOptions{Cwd: dir, Args: []string{"true"}}); err != nil {
		t.Fatal(err)
	}
	m := latestManifest(t, dir)

	raw, err := os.ReadFile(filepath.Join(dir, "runs", m.RunID, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}

	want := map[string]bool{
		"run_id": true, "project": true, "git_commit": true, "git_branch": true,
		"git_dirty": true, "status": true, "started_at": true, "duration_seconds": true,
		"command": true, "host": true, "exit_code": true, "stdout_log": true,
		"stderr_log": true, "parameters": true, "metrics": true,
	}
	for k := range want {
		if _, ok := got[k]; !ok {
			t.Errorf("missing key %q", k)
		}
	}
	if len(got) != len(want) {
		t.Errorf("got %d keys, want %d — extra keys mean schema drift", len(got), len(want))
	}
}
