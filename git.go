package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

func GetCurrentTag() (string, error) {
	cmd := exec.Command("git", "describe", "--tags", "--abbrev=0", "--match", "v*.*.*")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("git: %s", &stderr)
	}

	return strings.TrimSpace(stdout.String()), nil
}
func HasTag() (bool, error) {
	cmd := exec.Command("git", "tag", "--list")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return false, fmt.Errorf("git: %s", &stderr)
	}

	scanner := bufio.NewScanner(&stdout)
	for scanner.Scan() {
		tag := strings.TrimSpace(scanner.Text())
		_, err := ParseVersion(tag)
		if err == nil {
			return true, nil
		}
	}
	if err := scanner.Err(); err != nil {
		return false, err
	}

	return false, nil
}

func AddVersionTag(v Version) error {
	cmd := exec.Command("git", "tag", v.String())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("git: %s", output)
	}
	return nil
}
