package utils

import (
	"fmt"
	"os"
	"os/exec"
)

func WriteToFile(content string, filePath string) error {

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func ExecuteCommand(command string, args []string) (string, error) {

	cmd := exec.Command(command, args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("command execution failed: %w\nOutput: %s", err, output)
	}

	return fmt.Sprintf("%s", output), nil
}
