package simtrack

import (
	"errors"
	"os/exec"
	"strings"
)

type gitState struct {
	Root   string
	Commit string
	Branch string
	Dirty  bool
}

// capture GitState and populate the gitState Struct
func CaptureGitState(dir string) (*gitState, error) {
	root, err := git(dir, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, errors.New("not inside a Git Repo")
	}

	commit, err := git(root, "rev-parse", "HEAD")
	if err != nil {
		return nil, errors.New("no-commits")
	}

	branch, err := git(root, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return nil, errors.New("unknown")
	}

	dirty := false
	if _, err := git(root, "diff", "--quiet"); err != nil {
		dirty = true
	}

	if _, err := git(root, "diff", "--cached", "--quiet"); err != nil {
		dirty = true
	}

	return &gitState{Root: root, Commit: commit, Branch: branch, Dirty: dirty}, nil
}

// Returns the Git root of the current directory, or an error if it cannot be determined
func FindProjectRoot() (string, error) {
	gs, err := CaptureGitState(".")
	if err != nil {
		return "", err
	}

	return gs.Root, nil
}

// The execution of the git command
func git(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.Output()
	return strings.TrimSpace(string(out)), err
}
