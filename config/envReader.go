package config

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func SetupEnvVariables(pathToEnvFile string) error {
	file, err := os.OpenFile(pathToEnvFile, os.O_RDONLY, 0o644)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", pathToEnvFile, err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++

		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid format at line %d: %s", lineNum, line)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return fmt.Errorf("empty key at line %d", lineNum)
		}
		err := os.Setenv(key, value)
		if err != nil {
			return err
		}

	}
	return nil
}
