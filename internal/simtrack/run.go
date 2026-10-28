package simtrack

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
	"errors"
)

type RunOptions struct {
	Cwd  string   // optional; defaults to current directory
	Args []string // command + arguments (post `--`)
}

// Run executes opts.Args, records a manifest, and returns the command's exit code.
func Run(opts RunOptions) (int, error) {
	if len(opts.Args) == 0 {
		return 0, fmt.Errorf("no command given")
	}
	if opts.Cwd == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return 0, err
		}
		opts.Cwd = cwd
	}

	gs, err := CaptureGitState(opts.Cwd)
	if err != nil {
		return 0, err
	}

	startedAt := time.Now().UTC()
	shortSHA := gs.Commit
	if len(shortSHA) > 7 {
		shortSHA = shortSHA[:7]
	}
	runID := fmt.Sprintf("run-%s-%s", startedAt.Format("20060102-150405"), shortSHA)
	runDir := filepath.Join(gs.Root, "runs", runID)
	if err := os.MkdirAll(filepath.Join(runDir, "artifacts"), 0o755); err != nil {
		return 0, err
	}

	host, _ := os.Hostname()
	command := strings.Join(opts.Args, " ")

	m := &Manifest{
		RunID:      runID,
		Project:    filepath.Base(gs.Root),
		GitCommit:  gs.Commit,
		GitBranch:  gs.Branch,
		GitDirty:   gs.Dirty,
		Status:     "running",
		StartedAt:  startedAt.Format(time.RFC3339),
		Command:    command,
		Host:       host,
		StdoutLog:  "stdout.log",
		StderrLog:  "stderr.log",
		Parameters: map[string]any{},
		Metrics:    map[string]any{},
	}
	manifestPath := filepath.Join(runDir, "manifest.json")
	if err := writeManifest(manifestPath, m); err != nil {
		return 0, err
	}

	fmt.Printf("[simtrack] run_id:  %s\n", runID)
	fmt.Printf("[simtrack] command: %s\n", command)
	fmt.Printf("[simtrack] running...\n")

	stdoutPath := filepath.Join(runDir, "stdout.log")
	stderrPath := filepath.Join(runDir, "stderr.log")
	outFile, err := os.Create(stdoutPath)
	if err != nil {
		return 0, err
	}
	errFile, err := os.Create(stderrPath)
	if err != nil {
		outFile.Close()
		return 0, err
	}

	execStart := time.Now()
	cmd := exec.Command(opts.Args[0], opts.Args[1:]...)
	cmd.Dir = gs.Root // match shell wrapper: always run from project root
	cmd.Stdout = outFile
	cmd.Stderr = errFile
	cmd.Stdin = os.Stdin
	runErr := cmd.Run()
	outFile.Close()
	errFile.Close()

	exitCode := 0
	if runErr != nil {
		var ee *exec.ExitError
		if errors.As(runErr, &ee) {
			exitCode = ee.ExitCode()
		} else {
			exitCode = 1
		}
	}
	durationSec := int(time.Since(execStart).Seconds())
	status := "completed"
	if exitCode != 0 {
		status = "failed"
	}

	m.Status = status
	m.ExitCode = &exitCode
	m.DurationSeconds = &durationSec
	if err := writeManifest(manifestPath, m); err != nil {
		return exitCode, err
	}

	fmt.Printf("[simtrack] status:  %s (exit=%d, duration=%ds)\n", status, exitCode, durationSec)
	fmt.Printf("[simtrack] run dir: %s\n", runDir)
	fmt.Printf("[simtrack] logs:    %s  %s\n", stdoutPath, stderrPath)
	return exitCode, nil
}
