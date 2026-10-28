package simtrack 

import (
	"errors"
	"os/exec"
	"strings"
)

type gitState struct {
	Root string 
	Commit string 
	Branch string 
	Dirty bool
}


// capture GitState and populate the gitState Struct 
func captureGitState(dir string) (*gitState, error) {

}
// Returns the Git root of the current directory, or an error if it cannot be determined
func FinProjectRoot() (string, error) {

}
// Returns the current Git commit hash, or an error if it cannot be determined
func git(dir string, args ...string) (string, error) {
}
