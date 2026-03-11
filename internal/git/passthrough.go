package git

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func ExecGit(args []string) error {
	gitPath, err := exec.LookPath("git")
	if err != nil {
		return fmt.Errorf("infra: git not found in PATH: %w", err)
	}

	gitArgs := append([]string{"git"}, args...)

	return syscall.Exec(gitPath, gitArgs, os.Environ())
}
