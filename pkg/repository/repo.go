package repository

import (
	"bytes"
	"gigolow/configs"
	"gigolow/pkg/logging"
	"os/exec"
	"strings"
)

func Status(repository configs.Repository, logger *logging.Logger) (string, error) {
	cmd := exec.Command("git", "status")
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	logger.Log("Executing git status command")

	err := cmd.Run()
	logger.Log(out.String())
	if err != nil {
		logger.Log("Error executing git status command for repository: " + repository.Url + ". Error: " + err.Error())
	}
	return out.String(), err
}
func Clone(repository configs.Repository, logger *logging.Logger) error {
	args := []string{"clone", repository.Url}
	if repository.Branch != "" {
		args = append(args, "-b", repository.Branch)
	}

	cmd := exec.Command("git", args...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	logger.Log("Executing git clone command: git " + strings.Join(args, " "))

	err := cmd.Run()
	logger.Log(out.String())
	if err != nil {
		logger.Log("Error executing git clone command for repository: " + repository.Url + ", branch: " + repository.Url + ". Error: " + err.Error())
	}
	return err
}
