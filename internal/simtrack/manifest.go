package simtrack 

import (
	"encoding/json"
	"os"
)

type Manifest struct {
	RunID           string         `json:"run_id"`
	Project         string         `json:"project"`
	GitCommit       string         `json:"git_commit"`
	GitBranch       string         `json:"git_branch"`
	GitDirty        bool           `json:"git_dirty"`
	Status          string         `json:"status"`
	StartedAt       string         `json:"started_at"`
	DurationSeconds *int           `json:"duration_seconds"`
	Command         string         `json:"command"`
	Host            string         `json:"host"`
	ExitCode        *int           `json:"exit_code"`
	StdoutLog       string         `json:"stdout_log"`
	StderrLog       string         `json:"stderr_log"`
	Parameters      map[string]any `json:"parameters"`
	Metrics         map[string]any `json:"metrics"`
}


func writeManifest(path string, m *Manifest) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	
	enc := json.NewEncoder(f)
	enc.SetIndent(""," ")
	return enc.Encode(m)

}

func readManifest(path string) (*Manifest, error) {
	f, e := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	var m Manifest 
	err = json.NewDecoder(file).Decode(&Manifest)
	if err != nil {
		return nil, err
	}
	return $m, nil
}
